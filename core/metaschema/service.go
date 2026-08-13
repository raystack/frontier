package metaschema

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/raystack/frontier/pkg/utils"

	"github.com/pkg/errors"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/robfig/cron/v3"
	"github.com/xeipuuv/gojsonschema"
)

type Service struct {
	repository      Repository
	logger          *slog.Logger
	refreshInterval time.Duration

	mu              sync.RWMutex
	metaSchemaCache map[string]MetaSchema

	syncJob   *cron.Cron
	syncJobMu sync.Mutex
}

func NewService(repository Repository, logger *slog.Logger, refreshInterval time.Duration) *Service {
	return &Service{
		repository:      repository,
		logger:          logger,
		refreshInterval: refreshInterval,
		metaSchemaCache: make(map[string]MetaSchema),
	}
}

func (s *Service) Create(ctx context.Context, toCreate MetaSchema) (MetaSchema, error) {
	mschema, err := s.repository.Create(ctx, toCreate)
	if err != nil {
		return MetaSchema{}, err
	}
	s.mu.Lock()
	s.metaSchemaCache[mschema.Name] = mschema
	s.mu.Unlock()
	return mschema, nil
}

func (s *Service) Get(ctx context.Context, idOrName string) (MetaSchema, error) {
	s.mu.RLock()
	schema, ok := s.metaSchemaCache[idOrName]
	s.mu.RUnlock()
	if ok {
		return schema, nil
	}

	if utils.IsValidUUID(idOrName) {
		schema, err := s.repository.Get(ctx, idOrName)
		if err != nil {
			return MetaSchema{}, err
		}
		return schema, nil
	}
	return MetaSchema{}, ErrInvalidID
}

func (s *Service) List(ctx context.Context) ([]MetaSchema, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	schemas := make([]MetaSchema, 0, len(s.metaSchemaCache))
	for _, schema := range s.metaSchemaCache {
		schemas = append(schemas, schema)
	}
	return schemas, nil
}

func (s *Service) Update(ctx context.Context, id string, toUpdate MetaSchema) (MetaSchema, error) {
	if utils.IsValidUUID(id) {
		schema, err := s.repository.Update(ctx, id, toUpdate)
		if err != nil {
			return MetaSchema{}, err
		}
		s.mu.Lock()
		s.metaSchemaCache[schema.Name] = schema
		s.mu.Unlock()
		return schema, nil
	}
	return MetaSchema{}, ErrInvalidID
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if utils.IsValidUUID(id) {
		name, err := s.repository.Delete(ctx, id)
		if err != nil {
			return err
		}
		s.mu.Lock()
		delete(s.metaSchemaCache, name)
		s.mu.Unlock()
		return nil
	}
	return ErrInvalidID
}

func (s *Service) MigrateDefault(ctx context.Context) error {
	return s.repository.MigrateDefaults(ctx)
}

// Validate checks the metadata against the json-schema. When the named
// metaschema is not in the cache it returns nil (no validation).
func (s *Service) Validate(mdata metadata.Metadata, name string) error {
	s.mu.RLock()
	mschema, ok := s.metaSchemaCache[name]
	s.mu.RUnlock()
	if !ok {
		return nil
	}

	metadataSchema := gojsonschema.NewStringLoader(mschema.Schema)
	providedSchema := gojsonschema.NewGoLoader(mdata)
	results, err := gojsonschema.Validate(metadataSchema, providedSchema)
	if err != nil {
		return errors.Wrap(err, "failed to validate metadata")
	}

	if !results.Valid() {
		return errors.New("metadata doesn't match the json-schema")
	}
	return nil
}

// reload replaces the cache with the current set of metaschemas from the
// database. On error it keeps the existing cache, so a database blip does not
// blank validation.
func (s *Service) reload(ctx context.Context) {
	schemas, err := s.repository.List(ctx)
	if err != nil {
		s.logger.WarnContext(ctx, "metaschema cache refresh failed", "err", err)
		return
	}
	fresh := make(map[string]MetaSchema, len(schemas))
	for _, schema := range schemas {
		fresh[schema.Name] = schema
	}
	s.mu.Lock()
	s.metaSchemaCache = fresh
	s.mu.Unlock()
}

// Init primes the cache from the database and, when refreshInterval is greater
// than zero, starts a background job that reloads it on that interval so a
// change made on one pod reaches the others.
func (s *Service) Init(ctx context.Context) error {
	s.reload(ctx)
	s.mu.RLock()
	count := len(s.metaSchemaCache)
	s.mu.RUnlock()
	s.logger.InfoContext(ctx, "metaschemas loaded", "count", count)

	if s.refreshInterval <= 0 {
		return nil
	}

	s.syncJobMu.Lock()
	defer s.syncJobMu.Unlock()
	if s.syncJob != nil {
		<-s.syncJob.Stop().Done()
	}
	s.syncJob = cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DefaultLogger),
		cron.Recover(cron.DefaultLogger),
	))
	if _, err := s.syncJob.AddFunc(fmt.Sprintf("@every %s", s.refreshInterval.String()), func() {
		s.reload(ctx)
	}); err != nil {
		return err
	}
	s.syncJob.Start()
	return nil
}

// Close stops the background refresh job. It is safe to call when none started.
func (s *Service) Close() error {
	s.syncJobMu.Lock()
	defer s.syncJobMu.Unlock()
	if s.syncJob != nil {
		<-s.syncJob.Stop().Done()
	}
	return nil
}
