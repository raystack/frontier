package audit

import (
	"context"
)

type Repository interface {
	Create(context.Context, *Log) error
	List(context.Context, Filter) ([]Log, error)
	GetByID(context.Context, string) (Log, error)
}

type Publisher interface {
	Publish(context.Context, Log)
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

type Service struct {
	source     string
	repository Repository
	publisher  Publisher

	actorExtractor    func(context.Context) (Actor, bool)
	metadataExtractor func(context.Context) (map[string]string, bool)
}

func NewService(source string, repository Repository, opts ...Option) *Service {
	svc := &Service{
		source:            source,
		repository:        repository,
		actorExtractor:    defaultActorExtractor,
		metadataExtractor: defaultMetadataExtractor,
	}
	for _, o := range opts {
		o(svc)
	}
	return svc
}

func (s *Service) Create(ctx context.Context, l *Log) error {
	err := s.repository.Create(ctx, l)
	if err != nil {
		return err
	}
	if s.publisher != nil {
		s.publisher.Publish(ctx, *l)
	}
	return nil
}

func (s *Service) List(ctx context.Context, flt Filter) ([]Log, error) {
	return s.repository.List(ctx, flt)
}

func (s *Service) GetByID(ctx context.Context, id string) (Log, error) {
	return s.repository.GetByID(ctx, id)
}
