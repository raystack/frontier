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

type fakeWebhookAPI struct {
	webhooks  []*frontierv1beta1.Webhook
	created   []*frontierv1beta1.WebhookRequestBody
	updated   map[string]*frontierv1beta1.WebhookRequestBody
	deleted   []string
	createErr error
}

func (f *fakeWebhookAPI) ListWebhooks(_ context.Context, _ *connect.Request[frontierv1beta1.ListWebhooksRequest]) (*connect.Response[frontierv1beta1.ListWebhooksResponse], error) {
	return connect.NewResponse(&frontierv1beta1.ListWebhooksResponse{Webhooks: f.webhooks}), nil
}

func (f *fakeWebhookAPI) CreateWebhook(_ context.Context, req *connect.Request[frontierv1beta1.CreateWebhookRequest]) (*connect.Response[frontierv1beta1.CreateWebhookResponse], error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.created = append(f.created, req.Msg.GetBody())
	return connect.NewResponse(&frontierv1beta1.CreateWebhookResponse{}), nil
}

func (f *fakeWebhookAPI) UpdateWebhook(_ context.Context, req *connect.Request[frontierv1beta1.UpdateWebhookRequest]) (*connect.Response[frontierv1beta1.UpdateWebhookResponse], error) {
	if f.updated == nil {
		f.updated = map[string]*frontierv1beta1.WebhookRequestBody{}
	}
	f.updated[req.Msg.GetId()] = req.Msg.GetBody()
	return connect.NewResponse(&frontierv1beta1.UpdateWebhookResponse{}), nil
}

func (f *fakeWebhookAPI) DeleteWebhook(_ context.Context, req *connect.Request[frontierv1beta1.DeleteWebhookRequest]) (*connect.Response[frontierv1beta1.DeleteWebhookResponse], error) {
	f.deleted = append(f.deleted, req.Msg.GetId())
	return connect.NewResponse(&frontierv1beta1.DeleteWebhookResponse{}), nil
}

func webhookPB(id, url, desc, state string, events ...string) *frontierv1beta1.Webhook {
	return &frontierv1beta1.Webhook{Id: id, Url: url, Description: desc, State: state, SubscribedEvents: events}
}

func TestWebhookReconciler(t *testing.T) {
	t.Run("applies add, update, and delete", func(t *testing.T) {
		api := &fakeWebhookAPI{webhooks: []*frontierv1beta1.Webhook{
			webhookPB("w1", "https://keep.example/hook", "old", webhookStateEnabled, "org.created"),
			webhookPB("w2", "https://gone.example/hook", "", webhookStateEnabled, "org.created"),
		}}
		spec := []byte(`
- {url: "https://new.example/hook", subscribed_events: [org.created]}
- {url: "https://keep.example/hook", description: updated, subscribed_events: [org.created]}
- {url: "https://gone.example/hook", delete: true}
`)
		rep, err := NewWebhookReconciler(api, "").Reconcile(context.Background(), spec, false)

		assert.NoError(t, err)
		assert.Equal(t, 3, rep.Applied)
		if assert.Len(t, api.created, 1) {
			assert.Equal(t, "https://new.example/hook", api.created[0].GetUrl())
			assert.Equal(t, []string{"org.created"}, api.created[0].GetSubscribedEvents())
		}
		if body := api.updated["w1"]; assert.NotNil(t, body) {
			assert.Equal(t, "https://keep.example/hook", body.GetUrl()) // identity never changes
			assert.Equal(t, "updated", body.GetDescription())
		}
		assert.Equal(t, []string{"w2"}, api.deleted)
	})

	t.Run("an update preserves headers and metadata it does not manage", func(t *testing.T) {
		md, _ := structpb.NewStruct(map[string]any{"team": "platform"})
		wh := webhookPB("w1", "https://a.example/hook", "desc", webhookStateEnabled, "org.created")
		wh.Headers = map[string]string{"X-Token": "abc"}
		wh.Metadata = md
		api := &fakeWebhookAPI{webhooks: []*frontierv1beta1.Webhook{wh}}
		// only the event set changes; headers and metadata must survive the update
		spec := []byte("- {url: \"https://a.example/hook\", subscribed_events: [org.created, org.deleted]}\n")

		rep, err := NewWebhookReconciler(api, "").Reconcile(context.Background(), spec, false)

		assert.NoError(t, err)
		assert.Equal(t, 1, rep.Applied)
		if body := api.updated["w1"]; assert.NotNil(t, body) {
			assert.Equal(t, map[string]string{"X-Token": "abc"}, body.GetHeaders())
			assert.Equal(t, "platform", body.GetMetadata().GetFields()["team"].GetStringValue())
			// state was not in the file; the update must carry the server value
			// through, not drop it to empty (which would reset it on the server).
			assert.Equal(t, webhookStateEnabled, body.GetState())
		}
	})

	t.Run("creates a webhook subscribed to all events when events are omitted", func(t *testing.T) {
		api := &fakeWebhookAPI{}
		spec := []byte("- {url: \"https://a.example/hook\"}\n")

		rep, err := NewWebhookReconciler(api, "").Reconcile(context.Background(), spec, false)

		assert.NoError(t, err)
		assert.Equal(t, []string{"add webhook https://a.example/hook [all events]"}, rep.Planned)
		if assert.Len(t, api.created, 1) {
			assert.Empty(t, api.created[0].GetSubscribedEvents())
		}
	})

	t.Run("an endpoint subscribed to all events round-trips", func(t *testing.T) {
		api := &fakeWebhookAPI{webhooks: []*frontierv1beta1.Webhook{
			webhookPB("w1", "https://a.example/hook", "", webhookStateEnabled), // no events = all events
		}}
		registry := map[string]Reconciler{KindWebhook: NewWebhookReconciler(api, "")}

		out, err := Export(context.Background(), registry, KindWebhook)
		assert.NoError(t, err)
		assert.Contains(t, string(out), "subscribed_events: []") // an all-events endpoint is written explicitly

		reports, err := Run(context.Background(), registry, out, true)
		assert.NoError(t, err)
		if assert.Len(t, reports, 1) {
			assert.Empty(t, reports[0].Planned)
		}
	})

	t.Run("a dry run plans without applying", func(t *testing.T) {
		api := &fakeWebhookAPI{}
		spec := []byte("- {url: \"https://a.example/hook\", subscribed_events: [org.created]}\n")

		rep, err := NewWebhookReconciler(api, "").Reconcile(context.Background(), spec, true)

		assert.NoError(t, err)
		assert.Equal(t, []string{"add webhook https://a.example/hook [org.created]"}, rep.Planned)
		assert.Zero(t, rep.Applied)
		assert.Empty(t, api.created)
	})

	t.Run("an unknown field in the spec fails the plan", func(t *testing.T) {
		api := &fakeWebhookAPI{}
		// `delet` instead of `delete`: must fail, not silently ignore it
		spec := []byte("- {url: \"https://a.example/hook\", subscribed_events: [org.created], delet: true}\n")

		_, err := NewWebhookReconciler(api, "").Reconcile(context.Background(), spec, true)

		assert.ErrorContains(t, err, "parse Webhook spec")
		assert.Empty(t, api.created)
	})

	t.Run("a create rejected by the server stops the run with the op named", func(t *testing.T) {
		api := &fakeWebhookAPI{createErr: connect.NewError(connect.CodeInvalidArgument, errors.New("bad request"))}
		spec := []byte("- {url: \"https://a.example/hook\", subscribed_events: [org.created]}\n")

		rep, err := NewWebhookReconciler(api, "").Reconcile(context.Background(), spec, false)

		assert.ErrorContains(t, err, "apply [add webhook https://a.example/hook [org.created]]")
		assert.Zero(t, rep.Applied)
		assert.Equal(t, []string{"add webhook https://a.example/hook [org.created]"}, rep.Planned)
	})

	t.Run("export round-trips to no changes and never leaks secrets", func(t *testing.T) {
		wh := webhookPB("w1", "https://a.example/hook", "desc", webhookStateDisabled, "org.created", "org.deleted")
		// even if a secret leaked into the response, the exporter must not carry it:
		// the spec has no secret field at all.
		wh.Secrets = []*frontierv1beta1.Webhook_Secret{{Id: "1", Value: "supersecret"}}
		api := &fakeWebhookAPI{webhooks: []*frontierv1beta1.Webhook{wh}}
		registry := map[string]Reconciler{KindWebhook: NewWebhookReconciler(api, "")}

		out, err := Export(context.Background(), registry, KindWebhook)
		assert.NoError(t, err)
		assert.Contains(t, string(out), "https://a.example/hook")
		assert.Contains(t, string(out), "disabled") // a non-default state is written
		assert.NotContains(t, string(out), "supersecret")

		reports, err := Run(context.Background(), registry, out, true)
		assert.NoError(t, err)
		if assert.Len(t, reports, 1) {
			assert.Empty(t, reports[0].Planned)
		}
	})

	t.Run("export omits the default state and round-trips", func(t *testing.T) {
		api := &fakeWebhookAPI{webhooks: []*frontierv1beta1.Webhook{
			webhookPB("w1", "https://a.example/hook", "", webhookStateEnabled, "org.created"),
		}}
		registry := map[string]Reconciler{KindWebhook: NewWebhookReconciler(api, "")}

		out, err := Export(context.Background(), registry, KindWebhook)
		assert.NoError(t, err)
		assert.NotContains(t, string(out), "state:") // enabled is the default, so it is left out

		reports, err := Run(context.Background(), registry, out, true)
		assert.NoError(t, err)
		if assert.Len(t, reports, 1) {
			assert.Empty(t, reports[0].Planned)
		}
	})

	t.Run("export with no endpoints yields an empty list", func(t *testing.T) {
		api := &fakeWebhookAPI{}
		registry := map[string]Reconciler{KindWebhook: NewWebhookReconciler(api, "")}

		out, err := Export(context.Background(), registry, KindWebhook)
		assert.NoError(t, err)
		assert.Contains(t, string(out), "spec: []")
	})
}
