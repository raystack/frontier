package invoice

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"
)

// TestService_InitClose_Concurrent guards against an unsynchronized syncJob field.
// Two goroutines is the minimum needed to surface the race under `go test -race`.
func TestService_InitClose_Concurrent(t *testing.T) {
	s := &Service{
		log:       slog.New(slog.NewTextHandler(io.Discard, nil)),
		syncDelay: time.Hour,
	}

	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := s.Init(context.Background()); err != nil {
				t.Errorf("Init: %v", err)
			}
		}()
	}
	wg.Wait()

	if err := s.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}
