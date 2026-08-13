package metaschema

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/raystack/frontier/pkg/metadata"
)

// fakeRepo is an in-memory metaschema.Repository for tests. It is safe for
// concurrent use so tests exercise the service's locking, not the repo's.
type fakeRepo struct {
	mu        sync.Mutex
	byID      map[string]MetaSchema
	listErr   error
	listCount int
	listHook  func() // called by List after it snapshots, with the repo mutex released
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
	f.listCount++
	if f.listErr != nil {
		err := f.listErr
		f.mu.Unlock()
		return nil, err
	}
	out := make([]MetaSchema, 0, len(f.byID))
	for _, m := range f.byID {
		out = append(out, m)
	}
	hook := f.listHook
	f.mu.Unlock()
	if hook != nil {
		hook()
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

func (f *fakeRepo) setListHook(fn func()) {
	f.mu.Lock()
	f.listHook = fn
	f.mu.Unlock()
}

func (f *fakeRepo) listCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.listCount
}

func (f *fakeRepo) clear() {
	f.mu.Lock()
	f.byID = map[string]MetaSchema{}
	f.mu.Unlock()
}

// waitForRow blocks until a row with the given name is committed, so a test can
// order a concurrent write ahead of the next step without a fixed sleep.
func (f *fakeRepo) waitForRow(t *testing.T, name string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		f.mu.Lock()
		found := false
		for _, m := range f.byID {
			if m.Name == name {
				found = true
				break
			}
		}
		f.mu.Unlock()
		if found {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("row %q never committed", name)
}

func discardLogger() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("condition not met within timeout")
}

func TestService_ConcurrentAccess(t *testing.T) {
	svc := NewService(newFakeRepo(), discardLogger(), 0)

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

func TestService_reload_picksUpNewSchema(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, discardLogger(), 0)
	if err := svc.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// Add a schema straight to the repo, as if another pod created it.
	if _, err := repo.Create(context.Background(), MetaSchema{ID: "1", Name: "user", Schema: `{"type":"object"}`}); err != nil {
		t.Fatalf("repo.Create: %v", err)
	}

	// Not visible yet: Init primed the cache before the row existed.
	if _, err := svc.Get(context.Background(), "user"); err == nil {
		t.Fatal("expected schema to be absent before reload")
	}

	if err := svc.reload(context.Background()); err != nil {
		t.Fatalf("reload: %v", err)
	}

	got, err := svc.Get(context.Background(), "user")
	if err != nil {
		t.Fatalf("Get after reload: %v", err)
	}
	if got.Name != "user" {
		t.Fatalf("got %q, want user", got.Name)
	}
}

func TestService_reload_keepsCacheOnListError(t *testing.T) {
	repo := newFakeRepo()
	_, _ = repo.Create(context.Background(), MetaSchema{ID: "1", Name: "user", Schema: `{"type":"object"}`})
	svc := NewService(repo, discardLogger(), 0)
	if err := svc.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	repo.listErr = context.DeadlineExceeded
	if err := svc.reload(context.Background()); err == nil {
		t.Fatal("reload should return the list error, not swallow it")
	}

	// The previous cache survives a failed reload.
	if _, err := svc.Get(context.Background(), "user"); err != nil {
		t.Fatalf("cache should survive a reload error: %v", err)
	}
}

// TestService_reload_keepsCacheOnEmptyList guards against an error-free empty
// read (a replica blip or a repo bug) blanking a populated cache.
func TestService_reload_keepsCacheOnEmptyList(t *testing.T) {
	repo := newFakeRepo()
	_, _ = repo.Create(context.Background(), MetaSchema{ID: "org", Name: "organization", Schema: `{"type":"object"}`})
	svc := NewService(repo, discardLogger(), 0)
	if err := svc.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	repo.clear() // List now returns an empty slice with no error

	if err := svc.reload(context.Background()); err == nil {
		t.Fatal("reload should report an error rather than blank a populated cache")
	}
	if _, err := svc.Get(context.Background(), "organization"); err != nil {
		t.Fatalf("cache should survive an empty list: %v", err)
	}
}

// TestService_reload_doesNotDropConcurrentWrite reproduces the read-modify-write
// clobber: reload snapshots the DB, a Create commits, then reload swaps. With a
// blind swap the created schema is lost; holding the write lock across the read
// and the swap keeps it.
func TestService_reload_doesNotDropConcurrentWrite(t *testing.T) {
	repo := newFakeRepo()
	// A base schema keeps the cache populated so the empty-list guard is not in play.
	_, _ = repo.Create(context.Background(), MetaSchema{ID: "org", Name: "organization", Schema: `{"type":"object"}`})
	svc := NewService(repo, discardLogger(), 0)
	if err := svc.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	snapshotted := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	repo.setListHook(func() {
		once.Do(func() {
			close(snapshotted)
			<-release
		})
	})

	reloadDone := make(chan struct{})
	go func() {
		_ = svc.reload(context.Background())
		close(reloadDone)
	}()

	<-snapshotted // reload has taken its DB snapshot, which does not include "user"

	createDone := make(chan struct{})
	go func() {
		_, _ = svc.Create(context.Background(), MetaSchema{ID: "user", Name: "user", Schema: `{"type":"object"}`})
		close(createDone)
	}()

	// Order the Create's row commit ahead of reload's swap: this is the window
	// the blind swap would clobber.
	repo.waitForRow(t, "user")
	close(release)

	<-reloadDone
	<-createDone

	if _, err := svc.Get(context.Background(), "user"); err != nil {
		t.Fatalf("concurrent Create was clobbered by reload: %v", err)
	}
}

func TestService_Init_refreshDisabled(t *testing.T) {
	svc := NewService(newFakeRepo(), discardLogger(), 0)
	if err := svc.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if svc.syncJob != nil {
		t.Fatal("interval 0 must not start a cron job")
	}
	if err := svc.Close(); err != nil {
		t.Fatalf("Close when no job started: %v", err)
	}
}

// TestService_Init_returnsPrimeError checks the boot-failure signal: a DB that
// is unreachable at startup must fail Init, not start with an empty cache.
func TestService_Init_returnsPrimeError(t *testing.T) {
	repo := newFakeRepo()
	repo.listErr = context.DeadlineExceeded
	svc := NewService(repo, discardLogger(), 0)
	if err := svc.Init(context.Background()); err == nil {
		t.Fatal("Init should return the initial prime error so startup can fail")
	}
}

// TestService_Init_runsScheduledRefresh exercises the cron lifecycle the PR
// adds: a positive interval starts the job, it refreshes on schedule, and Close
// stops it.
func TestService_Init_runsScheduledRefresh(t *testing.T) {
	repo := newFakeRepo()
	_, _ = repo.Create(context.Background(), MetaSchema{ID: "org", Name: "organization", Schema: `{"type":"object"}`})
	svc := NewService(repo, discardLogger(), 40*time.Millisecond)
	if err := svc.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}
	t.Cleanup(func() { _ = svc.Close() })

	if svc.syncJob == nil {
		t.Fatal("a positive interval must start a cron job")
	}

	// Prime is one List; wait for at least one scheduled refresh on top of it.
	waitFor(t, 2*time.Second, func() bool { return repo.listCalls() >= 2 })

	if err := svc.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	after := repo.listCalls()
	time.Sleep(120 * time.Millisecond)
	if got := repo.listCalls(); got != after {
		t.Fatalf("cron kept running after Close: was %d, now %d", after, got)
	}
}
