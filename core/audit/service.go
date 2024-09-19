package audit

import (
	"context"

	"golang.org/x/exp/slices"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/webhook"
)

type Repository interface {
	Create(context.Context, *Log) error
	List(context.Context, Filter) ([]Log, error)
	GetByID(context.Context, string) (Log, error)
}

type Publisher interface {
	Publish(context.Context, Log)
}

type WebhookService interface {
	Publish(ctx context.Context, e webhook.Event) error
}

type Option func(*Service)

func WithMetadataExtractor(fn func(context.Context) (map[string]string, bool)) Option {
	return func(s *Service) {
		s.metadataExtractor = fn
	}
}

func WithActorExtractor(fn func(context.Context) (Actor, bool)) Option {
	return func(s *Service) {
		s.actorExtractor = fn
	}
}

func WithLogPublisher(p Publisher) Option {
	return func(s *Service) {
		s.publisher = p
	}
}

func WithIgnoreList(items []string) Option {
	return func(s *Service) {
		s.ignoreList = items
	}
}

type Service struct {
	source         string
	repository     Repository
	publisher      Publisher
	webhookService WebhookService

	ignoreList        []string
	actorExtractor    func(context.Context) (Actor, bool)
	metadataExtractor func(context.Context) (map[string]string, bool)
}

func NewService(source string, repository Repository, webhookService WebhookService, opts ...Option) *Service {
	svc := &Service{
		source:            source,
		repository:        repository,
		actorExtractor:    defaultActorExtractor,
		metadataExtractor: defaultMetadataExtractor,
		webhookService:    webhookService,
	}
	for _, o := range opts {
		o(svc)
	}
	return svc
}

func (s *Service) Create(ctx context.Context, l *Log) error {
	if l.ID == "" {
		l.ID = uuid.NewString()
	}
	err := s.repository.Create(ctx, l)
	if err != nil {
		return err
	}

	if s.publisher != nil {
		if !slices.Contains(s.ignoreList, l.Action) {
			s.publisher.Publish(ctx, *l)
		}
	}
	if err := s.webhookService.Publish(ctx, webhook.Event{
		ID:        l.ID,
		Action:    l.Action,
		Data:      TransformToEventData(l),
		CreatedAt: l.CreatedAt,
	}); err != nil {
		return err
	}
	return nil
}

func (s *Service) List(ctx context.Context, flt Filter) ([]Log, error) {
	return s.repository.List(ctx, flt)
}

func (s *Service) GetByID(ctx context.Context, id string) (Log, error) {
	return s.repository.GetByID(ctx, id)
}

func TransformToEventData(l *Log) map[string]interface{} {
	anyMap := make(map[string]any)
	for k, v := range l.Metadata {
		anyMap[k] = v
	}
	result := map[string]any{
		"target": map[string]any{},
		"actor":  map[string]any{},
	}
	if l.Target.Name != "" {
		result["target"].(map[string]any)["name"] = l.Target.Name
	}
	if l.Target.ID != "" {
		result["target"].(map[string]any)["id"] = l.Target.ID
	}
	if l.Target.Type != "" {
		result["target"].(map[string]any)["type"] = l.Target.Type
	}
	if l.Actor.Name != "" {
		result["actor"].(map[string]any)["name"] = l.Actor.Name
	}
	if l.Actor.ID != "" {
		result["actor"].(map[string]any)["id"] = l.Actor.ID
	}
	if l.Actor.Type != "" {
		result["actor"].(map[string]any)["type"] = l.Actor.Type
	}
	if l.Source != "" {
		result["source"] = l.Source
	}
	if l.OrgID != "" {
		result["org_id"] = l.OrgID
	}
	result["metadata"] = anyMap
	return result
}
