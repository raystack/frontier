package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/webhook"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *ConnectHandler) CreateWebhook(ctx context.Context, req *connect.Request[frontierv1beta1.CreateWebhookRequest]) (*connect.Response[frontierv1beta1.CreateWebhookResponse], error) {
	errorLogger := NewErrorLogger()

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
		errorLogger.LogUnexpectedError(ctx, req, "CreateWebhook", err,
			zap.String("url", req.Msg.GetBody().GetUrl()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	endpointPb, err := toProtoWebhookEndpoint(endpoint)
	if err != nil {
		errorLogger.LogTransformError(ctx, req, "CreateWebhook", endpoint.ID, err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return connect.NewResponse(&frontierv1beta1.CreateWebhookResponse{
		Webhook: endpointPb,
	}), nil
}

func (h *ConnectHandler) UpdateWebhook(ctx context.Context, req *connect.Request[frontierv1beta1.UpdateWebhookRequest]) (*connect.Response[frontierv1beta1.UpdateWebhookResponse], error) {
	errorLogger := NewErrorLogger()
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
		errorLogger.LogUnexpectedError(ctx, req, "UpdateWebhook", err,
			zap.String("webhook_id", webhookID),
			zap.String("url", req.Msg.GetBody().GetUrl()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	endpointPb, err := toProtoWebhookEndpoint(endpoint)
	if err != nil {
		errorLogger.LogTransformError(ctx, req, "UpdateWebhook", endpoint.ID, err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return connect.NewResponse(&frontierv1beta1.UpdateWebhookResponse{
		Webhook: endpointPb,
	}), nil
}

func (h *ConnectHandler) ListWebhooks(ctx context.Context, req *connect.Request[frontierv1beta1.ListWebhooksRequest]) (*connect.Response[frontierv1beta1.ListWebhooksResponse], error) {
	errorLogger := NewErrorLogger()

	filter := webhook.EndpointFilter{}
	endpoints, err := h.webhookService.ListEndpoints(ctx, filter)
	if err != nil {
		errorLogger.LogUnexpectedError(ctx, req, "ListWebhooks", err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	var webhooks []*frontierv1beta1.Webhook
	for _, endpoint := range endpoints {
		endpointPb, err := toProtoWebhookEndpoint(endpoint)
		if err != nil {
			errorLogger.LogTransformError(ctx, req, "ListWebhooks", endpoint.ID, err)
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		webhooks = append(webhooks, endpointPb)
	}
	return connect.NewResponse(&frontierv1beta1.ListWebhooksResponse{
		Webhooks: webhooks,
	}), nil
}

func (h *ConnectHandler) DeleteWebhook(ctx context.Context, req *connect.Request[frontierv1beta1.DeleteWebhookRequest]) (*connect.Response[frontierv1beta1.DeleteWebhookResponse], error) {
	errorLogger := NewErrorLogger()
	webhookID := req.Msg.GetId()

	err := h.webhookService.DeleteEndpoint(ctx, webhookID)
	if err != nil {
		errorLogger.LogUnexpectedError(ctx, req, "DeleteWebhook", err,
			zap.String("webhook_id", webhookID))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
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
