package audit

import (
	"context"

	"github.com/raystack/frontier/core/webhook"
)

type NoopRepository struct{}

func NewNoopRepository() *NoopRepository {
	return &NoopRepository{}
}

func (r NoopRepository) Create(ctx context.Context, log *Log) error {
	return nil
}

func (r NoopRepository) List(ctx context.Context, filter Filter) ([]Log, error) {
	return []Log{}, nil
}

func (r NoopRepository) GetByID(ctx context.Context, s string) (Log, error) {
	return Log{}, nil
}

type NoopWebhookService struct{}

func NewNoopWebhookService() *NoopWebhookService {
	return &NoopWebhookService{}
}

func (s NoopWebhookService) Publish(ctx context.Context, e webhook.Event) error {
	return nil
}
