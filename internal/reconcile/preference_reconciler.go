package reconcile

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"connectrpc.com/connect"
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
	defaults, err := r.fetchDefaults(ctx)
	if err != nil {
		return Report{}, err
	}

	ops, err := diffPreferences(specs, current, defaults)
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
	defaults, err := r.fetchDefaults(ctx)
	if err != nil {
		return nil, err
	}

	specs := make([]PreferenceSpec, 0, len(current))
	for name, value := range current {
		def, known := defaults[name]
		if !known || value == def {
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

// fetchDefaults reads the platform traits and their defaults. It is both the
// map of defaults and the set of valid platform preference names.
func (r *PreferenceReconciler) fetchDefaults(ctx context.Context) (map[string]string, error) {
	resp, err := r.client.DescribePreferences(ctx, authReq(&frontierv1beta1.DescribePreferencesRequest{}, r.header))
	if err != nil {
		return nil, fmt.Errorf("describe preferences: %w", err)
	}
	defaults := map[string]string{}
	for _, t := range resp.Msg.GetTraits() {
		if t.GetResourceType() != schema.PlatformNamespace {
			continue
		}
		defaults[t.GetName()] = t.GetDefault()
	}
	return defaults, nil
}

func (r *PreferenceReconciler) apply(ctx context.Context, op preferenceOp) error {
	_, err := r.client.CreatePreferences(ctx, authReq(&frontierv1beta1.CreatePreferencesRequest{
		Preferences: []*frontierv1beta1.PreferenceRequestBody{{Name: op.name, Value: op.value}},
	}, r.header))
	return err
}
