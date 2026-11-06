package reconcile

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

type fakeBillingFeatureAPI struct {
	features  []*frontierv1beta1.Feature
	listErr   error
	updateErr error
	created   []*frontierv1beta1.CreateFeatureRequest
	updated   []*frontierv1beta1.UpdateFeatureRequest
}

func (f *fakeBillingFeatureAPI) ListFeatures(_ context.Context, _ *connect.Request[frontierv1beta1.ListFeaturesRequest]) (*connect.Response[frontierv1beta1.ListFeaturesResponse], error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return connect.NewResponse(&frontierv1beta1.ListFeaturesResponse{Features: f.features}), nil
}

func (f *fakeBillingFeatureAPI) CreateFeature(_ context.Context, req *connect.Request[frontierv1beta1.CreateFeatureRequest]) (*connect.Response[frontierv1beta1.CreateFeatureResponse], error) {
	f.created = append(f.created, req.Msg)
	return connect.NewResponse(&frontierv1beta1.CreateFeatureResponse{}), nil
}

func (f *fakeBillingFeatureAPI) UpdateFeature(_ context.Context, req *connect.Request[frontierv1beta1.UpdateFeatureRequest]) (*connect.Response[frontierv1beta1.UpdateFeatureResponse], error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	f.updated = append(f.updated, req.Msg)
	return connect.NewResponse(&frontierv1beta1.UpdateFeatureResponse{}), nil
}

func featurePB(t *testing.T, id, name, title string, metadata map[string]any, productIDs ...string) *frontierv1beta1.Feature {
	t.Helper()
	md, err := structpb.NewStruct(metadata)
	assert.NoError(t, err)
	return &frontierv1beta1.Feature{Id: id, Name: name, Title: title, Metadata: md, ProductIds: productIDs}
}

func TestBillingFeatureReconciler(t *testing.T) {
	t.Run("an update keeps the current name and product links and sends the desired title and metadata", func(t *testing.T) {
		api := &fakeBillingFeatureAPI{features: []*frontierv1beta1.Feature{
			featurePB(t, "f1", "audit_logs", "Audit Logs", nil, "p1", "p2"),
		}}
		spec := []byte(`
- name: audit_logs
  title: Audit logs
  metadata:
    info: "See [the docs](https://docs.example.com)."
`)
		rep, err := NewBillingFeatureReconciler(api, "").Reconcile(context.Background(), spec, false)

		assert.NoError(t, err)
		assert.Equal(t, []string{"update feature audit_logs (title, metadata)"}, rep.Planned)
		assert.Equal(t, 1, rep.Applied)
		if assert.Len(t, api.updated, 1) {
			req := api.updated[0]
			assert.Equal(t, "f1", req.GetId())
			assert.Equal(t, "audit_logs", req.GetBody().GetName())
			assert.Equal(t, []string{"p1", "p2"}, req.GetBody().GetProductIds())
			assert.Equal(t, "Audit logs", req.GetBody().GetTitle())
			assert.Equal(t, map[string]any{"info": "See [the docs](https://docs.example.com)."}, req.GetBody().GetMetadata().AsMap())
		}
		assert.Empty(t, api.created)
	})

	t.Run("an update with omitted title and metadata sends them empty", func(t *testing.T) {
		api := &fakeBillingFeatureAPI{features: []*frontierv1beta1.Feature{
			featurePB(t, "f1", "audit_logs", "Audit Logs", map[string]any{"info": "old"}, "p1"),
		}}
		_, err := NewBillingFeatureReconciler(api, "").Reconcile(context.Background(), []byte("- name: audit_logs\n"), false)

		assert.NoError(t, err)
		if assert.Len(t, api.updated, 1) {
			assert.Equal(t, "", api.updated[0].GetBody().GetTitle())
			assert.Empty(t, api.updated[0].GetBody().GetMetadata().AsMap())
			assert.Equal(t, []string{"p1"}, api.updated[0].GetBody().GetProductIds())
		}
	})

	t.Run("an update of a feature with no product links sends none", func(t *testing.T) {
		api := &fakeBillingFeatureAPI{features: []*frontierv1beta1.Feature{
			featurePB(t, "f1", "audit_logs", "Audit Logs", nil),
		}}
		_, err := NewBillingFeatureReconciler(api, "").Reconcile(context.Background(), []byte("- {name: audit_logs, title: Logs}\n"), false)

		assert.NoError(t, err)
		if assert.Len(t, api.updated, 1) {
			assert.Empty(t, api.updated[0].GetBody().GetProductIds())
		}
	})

	t.Run("a server feature with a name shorter than three characters is out of scope", func(t *testing.T) {
		api := &fakeBillingFeatureAPI{features: []*frontierv1beta1.Feature{
			featurePB(t, "f1", "audit_logs", "Audit Logs", nil),
			featurePB(t, "f2", "ab", "", nil),
		}}
		rep, err := NewBillingFeatureReconciler(api, "").Reconcile(context.Background(), []byte("- {name: audit_logs, title: Audit Logs}\n"), true)

		assert.NoError(t, err)
		assert.Empty(t, rep.Planned)
	})

	t.Run("a file entry with the exact name of a short server feature fails as out of scope", func(t *testing.T) {
		api := &fakeBillingFeatureAPI{features: []*frontierv1beta1.Feature{
			featurePB(t, "f2", "ab", "", nil),
		}}
		_, err := NewBillingFeatureReconciler(api, "").Reconcile(context.Background(), []byte("- name: ab\n"), true)

		assert.ErrorContains(t, err, `feature "ab" is out of scope`)
	})

	t.Run("a short new name that differs from a short server feature only by case is not out of scope", func(t *testing.T) {
		api := &fakeBillingFeatureAPI{features: []*frontierv1beta1.Feature{
			featurePB(t, "f2", "ab", "", nil),
		}}
		_, err := NewBillingFeatureReconciler(api, "").Reconcile(context.Background(), []byte("- name: AB\n"), true)

		if assert.Error(t, err) {
			assert.NotContains(t, err.Error(), "out of scope")
			assert.Contains(t, err.Error(), "at least 3 characters")
		}
	})

	t.Run("a server feature with unset metadata matches an entry without metadata", func(t *testing.T) {
		api := &fakeBillingFeatureAPI{features: []*frontierv1beta1.Feature{
			{Id: "f1", Name: "audit_logs", Title: "Audit Logs"},
		}}
		rep, err := NewBillingFeatureReconciler(api, "").Reconcile(context.Background(), []byte("- {name: audit_logs, title: Audit Logs}\n"), true)

		assert.NoError(t, err)
		assert.Empty(t, rep.Planned)
	})

	t.Run("a create sends no product ids", func(t *testing.T) {
		api := &fakeBillingFeatureAPI{}
		spec := []byte("- {name: co_pilot, title: Co-pilot, metadata: {info: help}}\n")

		rep, err := NewBillingFeatureReconciler(api, "").Reconcile(context.Background(), spec, false)

		assert.NoError(t, err)
		assert.Equal(t, []string{"add feature co_pilot"}, rep.Planned)
		if assert.Len(t, api.created, 1) {
			body := api.created[0].GetBody()
			assert.Equal(t, "co_pilot", body.GetName())
			assert.Equal(t, "Co-pilot", body.GetTitle())
			assert.Equal(t, map[string]any{"info": "help"}, body.GetMetadata().AsMap())
			assert.Empty(t, body.GetProductIds())
		}
	})

	t.Run("a dry run makes no writes", func(t *testing.T) {
		api := &fakeBillingFeatureAPI{features: []*frontierv1beta1.Feature{
			featurePB(t, "f1", "audit_logs", "Audit Logs", nil, "p1"),
		}}
		spec := []byte("- {name: audit_logs, title: Logs}\n- {name: co_pilot}\n")

		rep, err := NewBillingFeatureReconciler(api, "").Reconcile(context.Background(), spec, true)

		assert.NoError(t, err)
		assert.True(t, rep.DryRun)
		assert.Equal(t, []string{"add feature co_pilot", "update feature audit_logs (title)"}, rep.Planned)
		assert.Zero(t, rep.Applied)
		assert.Empty(t, api.created)
		assert.Empty(t, api.updated)
	})

	t.Run("a short new name in the file fails API validation", func(t *testing.T) {
		api := &fakeBillingFeatureAPI{}
		_, err := NewBillingFeatureReconciler(api, "").Reconcile(context.Background(), []byte("- name: xy\n"), true)
		assert.ErrorContains(t, err, "at least 3 characters")
	})

	t.Run("wraps a list error", func(t *testing.T) {
		api := &fakeBillingFeatureAPI{listErr: errors.New("boom")}
		_, err := NewBillingFeatureReconciler(api, "").Reconcile(context.Background(), []byte("- name: audit_logs\n"), true)
		assert.ErrorContains(t, err, "list features: boom")
	})

	t.Run("an apply error returns the partial report", func(t *testing.T) {
		api := &fakeBillingFeatureAPI{
			features: []*frontierv1beta1.Feature{
				featurePB(t, "f1", "audit_logs", "Audit Logs", nil, "p1"),
			},
			updateErr: errors.New("boom"),
		}
		spec := []byte("- {name: audit_logs, title: Logs}\n- {name: co_pilot}\n")

		rep, err := NewBillingFeatureReconciler(api, "").Reconcile(context.Background(), spec, false)

		assert.ErrorContains(t, err, "apply [update feature audit_logs (title)]: boom")
		assert.Equal(t, 1, rep.Applied)
		assert.Len(t, api.created, 1)
	})
}

func TestBillingFeatureReconcilerValidate(t *testing.T) {
	rec := NewBillingFeatureReconciler(&fakeBillingFeatureAPI{}, "")

	assert.NoError(t, rec.Validate([]byte("- {name: audit_logs, title: Audit Logs, metadata: {info: help, rank: 5}}\n")))

	cases := []struct {
		name    string
		spec    string
		wantErr string
	}{
		{"unknown field", "- {name: audit_logs, titel: Audit Logs}\n", "field titel not found"},
		{"metadata with a non-string nested key", "- {name: audit_logs, metadata: {limits: {1: one}}}\n", "metadata cannot be stored"},
		{"metadata with a timestamp", "- {name: audit_logs, metadata: {since: 2020-01-01}}\n", "metadata cannot be stored"},
		{"delete flag", "- {name: audit_logs, delete: true}\n", "cannot be deleted"},
		{"empty name", "- {title: Audit Logs}\n", "name is required"},
		{"duplicate name", "- {name: audit_logs}\n- {name: audit_logs}\n", "listed more than once"},
		{"name too short", "- {name: ab}\n", "at least 3 characters"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.ErrorContains(t, rec.Validate([]byte(c.spec)), c.wantErr)
		})
	}
}
