package reconcile

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/preference"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

// PreferenceAPI is the API subset the preference reconciler needs. The reads
// live on different services (platform preferences on AdminService, the trait
// list on FrontierService); the caller provides one value that serves both.
type PreferenceAPI interface {
	ListPreferences(context.Context, *connect.Request[frontierv1beta1.ListPreferencesRequest]) (*connect.Response[frontierv1beta1.ListPreferencesResponse], error)
	DescribePreferences(context.Context, *connect.Request[frontierv1beta1.DescribePreferencesRequest]) (*connect.Response[frontierv1beta1.DescribePreferencesResponse], error)
	CreatePreferences(context.Context, *connect.Request[frontierv1beta1.CreatePreferencesRequest]) (*connect.Response[frontierv1beta1.CreatePreferencesResponse], error)
}

// PreferenceReconciler makes platform preferences match the desired spec. A
// preference is a name and a value; a missing entry resets to the trait
// default, because settings always have a default to fall back to.
type PreferenceReconciler struct {
	client PreferenceAPI
	header string
}

func NewPreferenceReconciler(client PreferenceAPI, header string) *PreferenceReconciler {
	return &PreferenceReconciler{client: client, header: header}
}

func (r *PreferenceReconciler) Kind() string { return KindPreference }

// Validate checks the server-free parts of every entry (name present, no
// duplicates) so a bad entry stops the whole file before anything applies. The
// unknown-name check needs the server's trait list, so it stays in the diff.
func (r *PreferenceReconciler) Validate(spec []byte) error {
	var specs []PreferenceSpec
	if err := decodeSpec(spec, &specs); err != nil {
		return fmt.Errorf("parse %s spec: %w", KindPreference, err)
	}
	seen := map[string]struct{}{}
	for _, s := range specs {
		if strings.TrimSpace(s.Name) == "" {
			return fmt.Errorf("preference name is required")
		}
		if _, dup := seen[s.Name]; dup {
			return fmt.Errorf("preference %q is listed more than once", s.Name)
		}
		seen[s.Name] = struct{}{}
	}
	return nil
}

func (r *PreferenceReconciler) Reconcile(ctx context.Context, spec []byte, dryRun bool) (Report, error) {
	var specs []PreferenceSpec
	if err := decodeSpec(spec, &specs); err != nil {
		return Report{}, fmt.Errorf("parse %s spec: %w", KindPreference, err)
	}

	current, err := r.fetchCurrent(ctx)
	if err != nil {
		return Report{}, err
	}
	traits, err := r.fetchTraits(ctx)
	if err != nil {
		return Report{}, err
	}

	ops, err := diffPreferences(specs, current, traits)
	if err != nil {
		return Report{}, err
	}

	rep := Report{Kind: KindPreference, DryRun: dryRun}
	for _, op := range ops {
		rep.Planned = append(rep.Planned, op.String())
	}
	if dryRun {
		return rep, nil
	}
	for _, op := range ops {
		if err := r.apply(ctx, op); err != nil {
			return rep, fmt.Errorf("apply [%s]: %w", op, err)
		}
		rep.Applied++
	}
	return rep, nil
}

// Export returns the platform preferences whose value differs from the trait
// default, sorted by name. Preferences at their default stay out of the file,
// so reconciling an export plans no changes.
func (r *PreferenceReconciler) Export(ctx context.Context) (any, error) {
	current, err := r.fetchCurrent(ctx)
	if err != nil {
		return nil, err
	}
	traits, err := r.fetchTraits(ctx)
	if err != nil {
		return nil, err
	}

	specs := make([]PreferenceSpec, 0, len(current))
	for name, value := range current {
		trait, known := traits[name]
		// A stored empty value counts as unset, so it is already at its default;
		// leave it out, matching how the diff resolves it. Exporting it would emit
		// a value the plan then rejects.
		if !known || value == "" || value == trait.Default {
			continue
		}
		specs = append(specs, PreferenceSpec{Name: name, Value: value})
	}
	sort.Slice(specs, func(i, j int) bool { return specs[i].Name < specs[j].Name })
	return specs, nil
}

// fetchCurrent reads the platform preferences that are stored on the server.
// A trait with no stored value does not appear here; it is at its default.
func (r *PreferenceReconciler) fetchCurrent(ctx context.Context) (map[string]string, error) {
	resp, err := r.client.ListPreferences(ctx, authReq(&frontierv1beta1.ListPreferencesRequest{}, r.header))
	if err != nil {
		return nil, fmt.Errorf("list preferences: %w", err)
	}
	current := map[string]string{}
	for _, p := range resp.Msg.GetPreferences() {
		current[p.GetName()] = p.GetValue()
	}
	return current, nil
}

// fetchTraits reads the platform traits, keyed by name. A trait carries its
// default and its validator, so this one map is the source of defaults, the set
// of valid platform preference names, and the rule a value must pass.
func (r *PreferenceReconciler) fetchTraits(ctx context.Context) (map[string]preference.Trait, error) {
	resp, err := r.client.DescribePreferences(ctx, authReq(&frontierv1beta1.DescribePreferencesRequest{}, r.header))
	if err != nil {
		return nil, fmt.Errorf("describe preferences: %w", err)
	}
	traits := map[string]preference.Trait{}
	for _, t := range resp.Msg.GetTraits() {
		if t.GetResourceType() != schema.PlatformNamespace {
			continue
		}
		traits[t.GetName()] = traitFromPB(t)
	}
	return traits, nil
}

// traitFromPB rebuilds the core trait from its wire form, so the plan can
// validate a value with the same validator the server applies on write.
func traitFromPB(t *frontierv1beta1.PreferenceTrait) preference.Trait {
	tr := preference.Trait{
		Name:       t.GetName(),
		Default:    t.GetDefault(),
		Input:      inputTypeFromPB(t.GetInputType()),
		InputHints: t.GetInputHints(),
	}
	for _, o := range t.GetInputOptions() {
		tr.InputOptions = append(tr.InputOptions, preference.InputHintOption{
			Name:        o.GetName(),
			Description: o.GetDescription(),
		})
	}
	return tr
}

// inputTypeFromPB maps the wire input type back to the core one, the inverse of
// transformPreferenceTraitToPB in the API layer. An unset or unknown type falls
// back to text, which the trait validator treats as "any non-empty value".
func inputTypeFromPB(it frontierv1beta1.PreferenceTrait_InputType) preference.TraitInput {
	switch it {
	case frontierv1beta1.PreferenceTrait_INPUT_TYPE_TEXTAREA:
		return preference.TraitInputTextarea
	case frontierv1beta1.PreferenceTrait_INPUT_TYPE_SELECT:
		return preference.TraitInputSelect
	case frontierv1beta1.PreferenceTrait_INPUT_TYPE_COMBOBOX:
		return preference.TraitInputCombobox
	case frontierv1beta1.PreferenceTrait_INPUT_TYPE_CHECKBOX:
		return preference.TraitInputCheckbox
	case frontierv1beta1.PreferenceTrait_INPUT_TYPE_MULTISELECT:
		return preference.TraitInputMultiselect
	case frontierv1beta1.PreferenceTrait_INPUT_TYPE_NUMBER:
		return preference.TraitInputNumber
	default:
		return preference.TraitInputText
	}
}

func (r *PreferenceReconciler) apply(ctx context.Context, op preferenceOp) error {
	_, err := r.client.CreatePreferences(ctx, authReq(&frontierv1beta1.CreatePreferencesRequest{
		Preferences: []*frontierv1beta1.PreferenceRequestBody{{Name: op.name, Value: op.value}},
	}, r.header))
	return err
}
