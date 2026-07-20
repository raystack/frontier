package v1beta1connect

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/webhook"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// webhookErrCode maps a webhook service error to the connect status code the
// caller should see, so a bad request reads as invalid-argument rather than an
// internal error.
func webhookErrCode(err error) connect.Code {
	switch {
	case errors.Is(err, webhook.ErrInvalidDetail), errors.Is(err, webhook.ErrInvalidUUID):
		return connect.CodeInvalidArgument
	case errors.Is(err, webhook.ErrConflict):
		return connect.CodeAlreadyExists
	case errors.Is(err, webhook.ErrNotFound):
		return connect.CodeNotFound
	default:
		return connect.CodeInternal
	}
}

func (h *ConnectHandler) CreateWebhook(ctx context.Context, req *connect.Request[frontierv1beta1.CreateWebhookRequest]) (*connect.Response[frontierv1beta1.CreateWebhookResponse], error) {
	var metaDataMap metadata.Metadata
	if req.Msg.GetBody().GetMetadata() != nil {
		metaDataMap = metadata.Build(req.Msg.GetBody().GetMetadata().AsMap())
	}
	endpoint, err := h.webhookService.CreateEndpoint(ctx, webhook.Endpoint{
		Description:      req.Msg.GetBody().GetDescription(),
		SubscribedEvents: req.Msg.GetBody().GetSubscribedEvents(),
		Headers:          req.Msg.GetBody().GetHeaders(),
		URL:              req.Msg.GetBody().GetUrl(),
		State:            webhook.State(req.Msg.GetBody().GetState()),
		Metadata:         metaDataMap,
	})
	if err != nil {
		return nil, connect.NewError(webhookErrCode(err), fmt.Errorf("CreateWebhook: url=%s: %w", req.Msg.GetBody().GetUrl(), err))
	}
	endpointPb, err := toProtoWebhookEndpoint(endpoint)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("CreateWebhook: entity_id=%s: %w", endpoint.ID, err))
	}
	return connect.NewResponse(&frontierv1beta1.CreateWebhookResponse{
		Webhook: endpointPb,
	}), nil
}

func (h *ConnectHandler) UpdateWebhook(ctx context.Context, req *connect.Request[frontierv1beta1.UpdateWebhookRequest]) (*connect.Response[frontierv1beta1.UpdateWebhookResponse], error) {
	webhookID := req.Msg.GetId()

	var metaDataMap metadata.Metadata
	if req.Msg.GetBody().GetMetadata() != nil {
		metaDataMap = metadata.Build(req.Msg.GetBody().GetMetadata().AsMap())
	}
	endpoint, err := h.webhookService.UpdateEndpoint(ctx, webhook.Endpoint{
		ID:               webhookID,
		Description:      req.Msg.GetBody().GetDescription(),
		SubscribedEvents: req.Msg.GetBody().GetSubscribedEvents(),
		Headers:          req.Msg.GetBody().GetHeaders(),
		URL:              req.Msg.GetBody().GetUrl(),
		State:            webhook.State(req.Msg.GetBody().GetState()),
		Metadata:         metaDataMap,
	})
	if err != nil {
		return nil, connect.NewError(webhookErrCode(err), fmt.Errorf("UpdateWebhook: webhook_id=%s url=%s: %w", webhookID, req.Msg.GetBody().GetUrl(), err))
	}
	endpointPb, err := toProtoWebhookEndpoint(endpoint)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("UpdateWebhook: entity_id=%s: %w", endpoint.ID, err))
	}
	return connect.NewResponse(&frontierv1beta1.UpdateWebhookResponse{
		Webhook: endpointPb,
	}), nil
}

func (h *ConnectHandler) ListWebhooks(ctx context.Context, req *connect.Request[frontierv1beta1.ListWebhooksRequest]) (*connect.Response[frontierv1beta1.ListWebhooksResponse], error) {
	filter := webhook.EndpointFilter{}
	endpoints, err := h.webhookService.ListEndpoints(ctx, filter)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListWebhooks: %w", err))
	}
	var webhooks []*frontierv1beta1.Webhook
	for _, endpoint := range endpoints {
		endpointPb, err := toProtoWebhookEndpoint(endpoint)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListWebhooks: entity_id=%s: %w", endpoint.ID, err))
		}
		webhooks = append(webhooks, endpointPb)
	}
	return connect.NewResponse(&frontierv1beta1.ListWebhooksResponse{
		Webhooks: webhooks,
	}), nil
}

func (h *ConnectHandler) DeleteWebhook(ctx context.Context, req *connect.Request[frontierv1beta1.DeleteWebhookRequest]) (*connect.Response[frontierv1beta1.DeleteWebhookResponse], error) {
	webhookID := req.Msg.GetId()

	err := h.webhookService.DeleteEndpoint(ctx, webhookID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("DeleteWebhook: webhook_id=%s: %w", webhookID, err))
	}
	return connect.NewResponse(&frontierv1beta1.DeleteWebhookResponse{}), nil
}

func toProtoWebhookEndpoint(endpoint webhook.Endpoint) (*frontierv1beta1.Webhook, error) {
	metaData, err := endpoint.Metadata.ToStructPB()
	if err != nil {
		return nil, err
	}

	var secrets []*frontierv1beta1.Webhook_Secret
	for _, secret := range endpoint.Secrets {
		secrets = append(secrets, &frontierv1beta1.Webhook_Secret{
			Id:    secret.ID,
			Value: secret.Value,
		})
	}
	return &frontierv1beta1.Webhook{
		Id:               endpoint.ID,
		Description:      endpoint.Description,
		SubscribedEvents: endpoint.SubscribedEvents,
		Headers:          endpoint.Headers,
		Url:              endpoint.URL,
		State:            string(endpoint.State),
		Metadata:         metaData,
		Secrets:          secrets,
		CreatedAt:        timestamppb.New(endpoint.CreatedAt),
		UpdatedAt:        timestamppb.New(endpoint.UpdatedAt),
	}, nil
}
