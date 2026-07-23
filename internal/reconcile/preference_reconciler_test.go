package reconcile

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
)

type fakePreferenceAPI struct {
	prefs   []*frontierv1beta1.Preference
	traits  []*frontierv1beta1.PreferenceTrait
	created []*frontierv1beta1.PreferenceRequestBody
}

func (f *fakePreferenceAPI) ListPreferences(_ context.Context, _ *connect.Request[frontierv1beta1.ListPreferencesRequest]) (*connect.Response[frontierv1beta1.ListPreferencesResponse], error) {
	return connect.NewResponse(&frontierv1beta1.ListPreferencesResponse{Preferences: f.prefs}), nil
}

func (f *fakePreferenceAPI) DescribePreferences(_ context.Context, _ *connect.Request[frontierv1beta1.DescribePreferencesRequest]) (*connect.Response[frontierv1beta1.DescribePreferencesResponse], error) {
	return connect.NewResponse(&frontierv1beta1.DescribePreferencesResponse{Traits: f.traits}), nil
}

func (f *fakePreferenceAPI) CreatePreferences(_ context.Context, req *connect.Request[frontierv1beta1.CreatePreferencesRequest]) (*connect.Response[frontierv1beta1.CreatePreferencesResponse], error) {
	f.created = append(f.created, req.Msg.GetPreferences()...)
	return connect.NewResponse(&frontierv1beta1.CreatePreferencesResponse{}), nil
}

func platformTraitPB(name, def string) *frontierv1beta1.PreferenceTrait {
	return &frontierv1beta1.PreferenceTrait{Name: name, Default: def, ResourceType: schema.PlatformNamespace}
}

func prefPB(name, value string) *frontierv1beta1.Preference {
	return &frontierv1beta1.Preference{Name: name, Value: value}
}

func platformTraits() []*frontierv1beta1.PreferenceTrait {
	return []*frontierv1beta1.PreferenceTrait{
		platformTraitPB("disable_orgs_on_create", "false"),
		platformTraitPB("invite_with_roles", "true"),
		// a non-platform trait: it must be ignored, so its name is never valid here
		{Name: "mail_otp", Default: "false", ResourceType: schema.OrganizationNamespace},
	}
}

func TestPreferenceReconciler(t *testing.T) {
	t.Run("applies a set and a reset", func(t *testing.T) {
		api := &fakePreferenceAPI{
			traits: platformTraits(),
			prefs: []*frontierv1beta1.Preference{
				prefPB("invite_with_roles", "false"), // stored, not listed -> reset to "true"
			},
		}
		spec := []byte("- {name: disable_orgs_on_create, value: \"true\"}\n")

		rep, err := NewPreferenceReconciler(api, "").Reconcile(context.Background(), spec, false)

		assert.NoError(t, err)
		assert.Equal(t, []string{
			"set preference disable_orgs_on_create = true",
			"reset preference invite_with_roles to default",
		}, rep.Planned)
		assert.Equal(t, 2, rep.Applied)

		got := map[string]string{}
		for _, b := range api.created {
			got[b.GetName()] = b.GetValue()
		}
		assert.Equal(t, map[string]string{"disable_orgs_on_create": "true", "invite_with_roles": "true"}, got)
	})

	t.Run("a dry run plans without applying", func(t *testing.T) {
		api := &fakePreferenceAPI{traits: platformTraits()}
		spec := []byte("- {name: disable_orgs_on_create, value: \"true\"}\n")

		rep, err := NewPreferenceReconciler(api, "").Reconcile(context.Background(), spec, true)

		assert.NoError(t, err)
		assert.Equal(t, []string{"set preference disable_orgs_on_create = true"}, rep.Planned)
		assert.Zero(t, rep.Applied)
		assert.Empty(t, api.created)
	})

	t.Run("an unknown preference in the file fails before applying", func(t *testing.T) {
		api := &fakePreferenceAPI{traits: platformTraits()}
		spec := []byte("- {name: mail_otp, value: \"true\"}\n") // org trait, not a platform one

		_, err := NewPreferenceReconciler(api, "").Reconcile(context.Background(), spec, false)

		assert.ErrorContains(t, err, `unknown platform preference "mail_otp"`)
		assert.Empty(t, api.created)
	})

	t.Run("a value the trait rejects fails before applying", func(t *testing.T) {
		api := &fakePreferenceAPI{traits: platformTraits()}
		spec := []byte("- {name: disable_orgs_on_create, value: \"\"}\n") // empty is never accepted

		_, err := NewPreferenceReconciler(api, "").Reconcile(context.Background(), spec, false)

		assert.ErrorContains(t, err, "not a valid value")
		assert.Empty(t, api.created)
	})

	t.Run("an unknown field in the spec fails the plan", func(t *testing.T) {
		api := &fakePreferenceAPI{traits: platformTraits()}
		// `valu` instead of `value`: must fail, not silently ignore the value
		spec := []byte("- {name: disable_orgs_on_create, valu: \"true\"}\n")

		_, err := NewPreferenceReconciler(api, "").Reconcile(context.Background(), spec, true)

		assert.ErrorContains(t, err, "parse Preference spec")
		assert.Empty(t, api.created)
	})

	t.Run("export returns only overrides, sorted, and round-trips", func(t *testing.T) {
		api := &fakePreferenceAPI{
			traits: platformTraits(),
			prefs: []*frontierv1beta1.Preference{
				prefPB("invite_with_roles", "true"),      // equals default: omitted
				prefPB("disable_orgs_on_create", "true"), // override: exported
			},
		}
		registry := map[string]Reconciler{KindPreference: NewPreferenceReconciler(api, "")}

		out, err := Export(context.Background(), registry, KindPreference)
		assert.NoError(t, err)
		assert.Contains(t, string(out), "disable_orgs_on_create")
		assert.NotContains(t, string(out), "invite_with_roles")

		reports, err := Run(context.Background(), registry, out, true)
		assert.NoError(t, err)
		if assert.Len(t, reports, 1) {
			assert.Empty(t, reports[0].Planned)
		}
	})

	t.Run("export with everything at default yields an empty list", func(t *testing.T) {
		api := &fakePreferenceAPI{traits: platformTraits()}
		registry := map[string]Reconciler{KindPreference: NewPreferenceReconciler(api, "")}

		out, err := Export(context.Background(), registry, KindPreference)
		assert.NoError(t, err)
		assert.Contains(t, string(out), "spec: []")
	})
}
