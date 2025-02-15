package webhook

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/exp/slices"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/raystack/frontier/pkg/crypt"
)

const (
	DefaultSecretID    = "1"
	SignatureHeader    = "X-Signature"
	EndpointRetryCount = 3
)

type EndpointRepository interface {
	Create(ctx context.Context, endpoint Endpoint) (Endpoint, error)
	UpdateByID(ctx context.Context, endpoint Endpoint) (Endpoint, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter EndpointFilter) ([]Endpoint, error)
}

type Service struct {
	eRepo EndpointRepository
}

func NewService(eRepo EndpointRepository) *Service {
	return &Service{eRepo: eRepo}
}

func (s Service) CreateEndpoint(ctx context.Context, endpoint Endpoint) (Endpoint, error) {
	if endpoint.ID == "" {
		endpoint.ID = uuid.NewString()
	}
	if endpoint.State == "" {
		endpoint.State = Enabled
	}

	// generate a random secret in hex
	secretHex, err := crypt.NewEncryptionKeyInHex()
	if err != nil {
		return Endpoint{}, err
	}
	endpoint.Secrets = append(endpoint.Secrets, Secret{
		ID:    DefaultSecretID,
		Value: secretHex,
	})
	return s.eRepo.Create(ctx, endpoint)
}

func (s Service) UpdateEndpoint(ctx context.Context, endpoint Endpoint) (Endpoint, error) {
	if endpoint.ID == "" {
		return Endpoint{}, ErrInvalidUUID
	}
	updated, err := s.eRepo.UpdateByID(ctx, endpoint)
	if err != nil {
		return Endpoint{}, err
	}
	updated.Secrets = nil
	return updated, nil
}

func (s Service) DeleteEndpoint(ctx context.Context, id string) error {
	return s.eRepo.Delete(ctx, id)
}

func (s Service) ListEndpoints(ctx context.Context, filter EndpointFilter) ([]Endpoint, error) {
	endpoints, err := s.eRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	for i := range endpoints {
		endpoints[i].Secrets = nil
	}
	return endpoints, nil
}

func (s Service) Publish(ctx context.Context, evt Event) error {
	logger := grpczap.Extract(ctx)
	data, err := structpb.NewStruct(evt.Data)
	if err != nil {
		logger.Error("failed to convert data to structpb", zap.Error(err))
		return fmt.Errorf("failed to convert data to structpb: %w", err)
	}
	event := &frontierv1beta1.WebhookEvent{
		Id:        evt.ID,
		Action:    evt.Action,
		Data:      data,
		CreatedAt: timestamppb.New(evt.CreatedAt),
	}

	payload, err := protojson.Marshal(event)
	if err != nil {
		logger.Error("failed to marshal event", zap.Error(err))
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// send event to endpoints
	go func() {
		detachedRepoContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		endpoints, err := s.eRepo.List(detachedRepoContext, EndpointFilter{
			State: Enabled,
		})
		if err != nil {
			logger.Error("failed to list endpoints", zap.Error(err))
			return
		}
		var errs []error
		for _, endpoint := range endpoints {
			if len(endpoint.SubscribedEvents) > 0 && !slices.Contains(endpoint.SubscribedEvents, event.GetAction()) {
				continue
			}
			if len(endpoint.Secrets) == 0 {
				errs = append(errs, fmt.Errorf("no secret found for endpoint: %s", endpoint.ID))
				continue
			}

			// just use the first secret to sign the payload for now
			secret := endpoint.Secrets[0]
			signature, err := crypt.GenerateHMACFromHex(payload, secret.Value)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to generate HMAC: %w", err))
				continue
			}

			requestHeaders := make(map[string]string)
			for k, v := range endpoint.Headers {
				requestHeaders[k] = v
			}
			requestHeaders[SignatureHeader] = signatureHeader(signature, secret.ID)
			if err := post(endpoint.URL, requestHeaders, payload); err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			logger.Error("failed to send events", zap.Errors("errs", errs))
		}
	}()
	return nil
}

func signatureHeader(val string, id string) string {
	return fmt.Sprintf("%s=%s", id, val)
}

func post(url string, headers map[string]string, payload []byte) error {
	// post event
	client := resty.New().
		SetRetryCount(EndpointRetryCount).
		SetRetryWaitTime(3 * time.Second).
		SetRetryMaxWaitTime(20 * time.Second).
		SetTimeout(3 * time.Second).
		SetHeaders(headers)

	_, err := client.R().SetBody(payload).Post(url)
	return err
}
