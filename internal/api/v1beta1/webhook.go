package v1beta1

import (
	"context"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/frontier/core/webhook"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type WebhookService interface {
	CreateEndpoint(ctx context.Context, endpoint webhook.Endpoint) (webhook.Endpoint, error)
	UpdateEndpoint(ctx context.Context, endpoint webhook.Endpoint) (webhook.Endpoint, error)
	DeleteEndpoint(ctx context.Context, id string) error
	ListEndpoints(ctx context.Context, filter webhook.EndpointFilter) ([]webhook.Endpoint, error)
}

func (h Handler) CreateWebhook(ctx context.Context, req *frontierv1beta1.CreateWebhookRequest) (*frontierv1beta1.CreateWebhookResponse, error) {
	logger := grpczap.Extract(ctx)

	var metaDataMap metadata.Metadata
	if req.GetBody().GetMetadata() != nil {
		metaDataMap = metadata.Build(req.GetBody().GetMetadata().AsMap())
	}
	endpoint, err := h.webhookService.CreateEndpoint(ctx, webhook.Endpoint{
		Description:      req.GetBody().GetDescription(),
		SubscribedEvents: req.GetBody().GetSubscribedEvents(),
		Headers:          req.GetBody().GetHeaders(),
		URL:              req.GetBody().GetUrl(),
		State:            webhook.State(req.GetBody().GetState()),
		Metadata:         metaDataMap,
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	endpointPb, err := toProtoWebhookEndpoint(endpoint)
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &frontierv1beta1.CreateWebhookResponse{
		Webhook: endpointPb,
	}, nil
}

func (h Handler) UpdateWebhook(ctx context.Context, req *frontierv1beta1.UpdateWebhookRequest) (*frontierv1beta1.UpdateWebhookResponse, error) {
	logger := grpczap.Extract(ctx)

	var metaDataMap metadata.Metadata
	if req.GetBody().GetMetadata() != nil {
		metaDataMap = metadata.Build(req.GetBody().GetMetadata().AsMap())
	}
	endpoint, err := h.webhookService.UpdateEndpoint(ctx, webhook.Endpoint{
		ID:               req.GetId(),
		Description:      req.GetBody().GetDescription(),
		SubscribedEvents: req.GetBody().GetSubscribedEvents(),
		Headers:          req.GetBody().GetHeaders(),
		URL:              req.GetBody().GetUrl(),
		State:            webhook.State(req.GetBody().GetState()),
		Metadata:         metaDataMap,
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	endpointPb, err := toProtoWebhookEndpoint(endpoint)
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &frontierv1beta1.UpdateWebhookResponse{
		Webhook: endpointPb,
	}, nil
}

func (h Handler) ListWebhooks(ctx context.Context, req *frontierv1beta1.ListWebhooksRequest) (*frontierv1beta1.ListWebhooksResponse, error) {
	logger := grpczap.Extract(ctx)

	filter := webhook.EndpointFilter{}
	endpoints, err := h.webhookService.ListEndpoints(ctx, filter)
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	var webhooks []*frontierv1beta1.Webhook
	for _, endpoint := range endpoints {
		endpointPb, err := toProtoWebhookEndpoint(endpoint)
		if err != nil {
			logger.Error(err.Error())
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		webhooks = append(webhooks, endpointPb)
	}
	return &frontierv1beta1.ListWebhooksResponse{
		Webhooks: webhooks,
	}, nil
}

func (h Handler) DeleteWebhook(ctx context.Context, req *frontierv1beta1.DeleteWebhookRequest) (*frontierv1beta1.DeleteWebhookResponse, error) {
	logger := grpczap.Extract(ctx)

	err := h.webhookService.DeleteEndpoint(ctx, req.GetId())
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &frontierv1beta1.DeleteWebhookResponse{}, nil
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
