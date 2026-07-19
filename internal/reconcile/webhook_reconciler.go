package reconcile

import (
	"context"
	"fmt"
	"sort"

	"connectrpc.com/connect"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/structpb"
)

// WebhookAPI is the API subset the webhook reconciler needs. Every webhook
// operation lives on the admin service.
type WebhookAPI interface {
	ListWebhooks(context.Context, *connect.Request[frontierv1beta1.ListWebhooksRequest]) (*connect.Response[frontierv1beta1.ListWebhooksResponse], error)
	CreateWebhook(context.Context, *connect.Request[frontierv1beta1.CreateWebhookRequest]) (*connect.Response[frontierv1beta1.CreateWebhookResponse], error)
	UpdateWebhook(context.Context, *connect.Request[frontierv1beta1.UpdateWebhookRequest]) (*connect.Response[frontierv1beta1.UpdateWebhookResponse], error)
	DeleteWebhook(context.Context, *connect.Request[frontierv1beta1.DeleteWebhookRequest]) (*connect.Response[frontierv1beta1.DeleteWebhookResponse], error)
}

// WebhookReconciler makes webhook endpoints match the desired spec. The URL is
// the identity; description, subscribed events, and state are the managed
// fields. A missing endpoint fails the plan, and deletion needs an explicit
// delete flag.
type WebhookReconciler struct {
	client WebhookAPI
	header string
}

func NewWebhookReconciler(client WebhookAPI, header string) *WebhookReconciler {
	return &WebhookReconciler{client: client, header: header}
}

func (r *WebhookReconciler) Kind() string { return KindWebhook }

func (r *WebhookReconciler) Reconcile(ctx context.Context, spec []byte, dryRun bool) (Report, error) {
	var specs []WebhookSpec
	if err := decodeSpec(spec, &specs); err != nil {
		return Report{}, fmt.Errorf("parse %s spec: %w", KindWebhook, err)
	}

	current, err := r.fetchCurrent(ctx)
	if err != nil {
		return Report{}, err
	}

	ops, err := diffWebhooks(specs, current)
	if err != nil {
		return Report{}, err
	}

	rep := Report{Kind: KindWebhook, DryRun: dryRun}
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

// Export returns the current webhook endpoints as a desired-state spec, sorted
// by url. State is written only when it is not the default ("enabled") and an
// empty description is left out, so an omitted field on reconcile keeps the
// server value and the export round-trips to no changes. Secrets are never read
// from the server, so they can never appear in an export.
func (r *WebhookReconciler) Export(ctx context.Context) (any, error) {
	current, err := r.fetchCurrent(ctx)
	if err != nil {
		return nil, err
	}
	sort.Slice(current, func(i, j int) bool { return current[i].URL < current[j].URL })

	specs := make([]WebhookSpec, 0, len(current))
	for _, c := range current {
		entry := WebhookSpec{
			URL:              c.URL,
			Description:      c.Description,
			SubscribedEvents: uniqueSorted(c.SubscribedEvents),
		}
		if c.State != "" && c.State != webhookStateEnabled {
			entry.State = c.State
		}
		specs = append(specs, entry)
	}
	return specs, nil
}

func (r *WebhookReconciler) fetchCurrent(ctx context.Context) ([]currentWebhook, error) {
	resp, err := r.client.ListWebhooks(ctx, authReq(&frontierv1beta1.ListWebhooksRequest{}, r.header))
	if err != nil {
		return nil, fmt.Errorf("list webhooks: %w", err)
	}
	var current []currentWebhook
	for _, w := range resp.Msg.GetWebhooks() {
		var md map[string]any
		if m := w.GetMetadata(); m != nil {
			md = m.AsMap()
		}
		current = append(current, currentWebhook{
			ID:               w.GetId(),
			URL:              w.GetUrl(),
			Description:      w.GetDescription(),
			SubscribedEvents: w.GetSubscribedEvents(),
			State:            w.GetState(),
			Headers:          w.GetHeaders(),
			Metadata:         md,
		})
	}
	return current, nil
}

func (r *WebhookReconciler) apply(ctx context.Context, op webhookOp) error {
	switch op.action {
	case opAdd:
		body, err := webhookBody(op.spec, nil, nil)
		if err != nil {
			return err
		}
		_, err = r.client.CreateWebhook(ctx, authReq(&frontierv1beta1.CreateWebhookRequest{Body: body}, r.header))
		return err
	case opUpdate:
		body, err := webhookBody(op.spec, op.headers, op.metadata)
		if err != nil {
			return err
		}
		_, err = r.client.UpdateWebhook(ctx, authReq(&frontierv1beta1.UpdateWebhookRequest{Id: op.id, Body: body}, r.header))
		return err
	case opRemove:
		_, err := r.client.DeleteWebhook(ctx, authReq(&frontierv1beta1.DeleteWebhookRequest{Id: op.id}, r.header))
		return err
	default:
		return fmt.Errorf("unknown op action %q", op.action)
	}
}

// webhookBody builds the request body for an add or update. On an update, the
// endpoint's current headers and metadata are carried through so an update that
// only changes a managed field does not wipe values set elsewhere. The signing
// secret is never sent: the server owns it and generates one on create.
func webhookBody(spec WebhookSpec, headers map[string]string, metadata map[string]any) (*frontierv1beta1.WebhookRequestBody, error) {
	var md *structpb.Struct
	if len(metadata) > 0 {
		var err error
		md, err = structpb.NewStruct(metadata)
		if err != nil {
			return nil, fmt.Errorf("build webhook metadata: %w", err)
		}
	}
	return &frontierv1beta1.WebhookRequestBody{
		Url:              spec.URL,
		Description:      spec.Description,
		SubscribedEvents: spec.SubscribedEvents,
		State:            spec.State,
		Headers:          headers,
		Metadata:         md,
	}, nil
}
