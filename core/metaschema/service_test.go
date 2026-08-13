package metaschema

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"

	"github.com/raystack/frontier/pkg/metadata"
)

// fakeRepo is an in-memory metaschema.Repository for tests. It is safe for
// concurrent use so tests exercise the service's locking, not the repo's.
type fakeRepo struct {
	mu      sync.Mutex
	byID    map[string]MetaSchema
	listErr error
}

func newFakeRepo() *fakeRepo { return &fakeRepo{byID: map[string]MetaSchema{}} }

func (f *fakeRepo) Create(_ context.Context, m MetaSchema) (MetaSchema, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if m.ID == "" {
		m.ID = m.Name
	}
	f.byID[m.ID] = m
	return m, nil
}

func (f *fakeRepo) Get(_ context.Context, id string) (MetaSchema, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, ok := f.byID[id]
	if !ok {
		return MetaSchema{}, ErrInvalidID
	}
	return m, nil
}

func (f *fakeRepo) Update(_ context.Context, id string, m MetaSchema) (MetaSchema, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	m.ID = id
	f.byID[id] = m
	return m, nil
}

func (f *fakeRepo) List(_ context.Context) ([]MetaSchema, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.listErr != nil {
		return nil, f.listErr
	}
	out := make([]MetaSchema, 0, len(f.byID))
	for _, m := range f.byID {
		out = append(out, m)
	}
	return out, nil
}

func (f *fakeRepo) Delete(_ context.Context, id string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	name := f.byID[id].Name
	delete(f.byID, id)
	return name, nil
}

func (f *fakeRepo) MigrateDefaults(_ context.Context) error { return nil }

func discardLogger() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func TestService_ConcurrentAccess(t *testing.T) {
	svc := NewService(newFakeRepo())

	const goroutines = 50
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := "s" + string(rune('0'+(i%5)))
			_, _ = svc.Create(context.Background(), MetaSchema{Name: name, Schema: `{"type":"object"}`})
			_ = svc.Validate(metadata.Metadata{"a": "b"}, name)
			_, _ = svc.List(context.Background())
			_, _ = svc.Get(context.Background(), name)
		}(i)
	}
	wg.Wait()
}
