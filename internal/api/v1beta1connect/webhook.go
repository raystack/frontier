package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/frontier/core/webhook"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type WebhookService interface {
	CreateEndpoint(ctx context.Context, endpoint webhook.Endpoint) (webhook.Endpoint, error)
	UpdateEndpoint(ctx context.Context, endpoint webhook.Endpoint) (webhook.Endpoint, error)
	DeleteEndpoint(ctx context.Context, id string) error
	ListEndpoints(ctx context.Context, filter webhook.EndpointFilter) ([]webhook.Endpoint, error)
}

func (h *ConnectHandler) CreateWebhook(ctx context.Context, req *connect.Request[frontierv1beta1.CreateWebhookRequest]) (*connect.Response[frontierv1beta1.CreateWebhookResponse], error) {
	logger := grpczap.Extract(ctx)

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
		logger.Error(err.Error())
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	endpointPb, err := toProtoWebhookEndpoint(endpoint)
	if err != nil {
		logger.Error(err.Error())
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return connect.NewResponse(&frontierv1beta1.CreateWebhookResponse{
		Webhook: endpointPb,
	}), nil
}

func (h *ConnectHandler) UpdateWebhook(ctx context.Context, req *connect.Request[frontierv1beta1.UpdateWebhookRequest]) (*connect.Response[frontierv1beta1.UpdateWebhookResponse], error) {
	logger := grpczap.Extract(ctx)

	var metaDataMap metadata.Metadata
	if req.Msg.GetBody().GetMetadata() != nil {
		metaDataMap = metadata.Build(req.Msg.GetBody().GetMetadata().AsMap())
	}
	endpoint, err := h.webhookService.UpdateEndpoint(ctx, webhook.Endpoint{
		ID:               req.Msg.GetId(),
		Description:      req.Msg.GetBody().GetDescription(),
		SubscribedEvents: req.Msg.GetBody().GetSubscribedEvents(),
		Headers:          req.Msg.GetBody().GetHeaders(),
		URL:              req.Msg.GetBody().GetUrl(),
		State:            webhook.State(req.Msg.GetBody().GetState()),
		Metadata:         metaDataMap,
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	endpointPb, err := toProtoWebhookEndpoint(endpoint)
	if err != nil {
		logger.Error(err.Error())
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return connect.NewResponse(&frontierv1beta1.UpdateWebhookResponse{
		Webhook: endpointPb,
	}), nil
}

func (h *ConnectHandler) ListWebhooks(ctx context.Context, req *connect.Request[frontierv1beta1.ListWebhooksRequest]) (*connect.Response[frontierv1beta1.ListWebhooksResponse], error) {
	logger := grpczap.Extract(ctx)

	filter := webhook.EndpointFilter{}
	endpoints, err := h.webhookService.ListEndpoints(ctx, filter)
	if err != nil {
		logger.Error(err.Error())
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	var webhooks []*frontierv1beta1.Webhook
	for _, endpoint := range endpoints {
		endpointPb, err := toProtoWebhookEndpoint(endpoint)
		if err != nil {
			logger.Error(err.Error())
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		webhooks = append(webhooks, endpointPb)
	}
	return connect.NewResponse(&frontierv1beta1.ListWebhooksResponse{
		Webhooks: webhooks,
	}), nil
}

func (h *ConnectHandler) DeleteWebhook(ctx context.Context, req *connect.Request[frontierv1beta1.DeleteWebhookRequest]) (*connect.Response[frontierv1beta1.DeleteWebhookResponse], error) {
	logger := grpczap.Extract(ctx)

	err := h.webhookService.DeleteEndpoint(ctx, req.Msg.GetId())
	if err != nil {
		logger.Error(err.Error())
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
