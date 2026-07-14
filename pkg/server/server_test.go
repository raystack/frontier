package server

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/raystack/frontier/internal/api"
)

func TestGracefulShutdownDrainsInflightRequests(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	entered := make(chan struct{})
	requestDone := make(chan struct{})
	mux := http.NewServeMux()
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		time.Sleep(300 * time.Millisecond)
		close(requestDone)
		w.WriteHeader(http.StatusOK)
	})

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := &http.Server{Handler: mux}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go gracefulShutdown(ctx, logger, &wg, srv, "test server", make(chan struct{}))

	serveReturned := make(chan struct{})
	go func() {
		defer close(serveReturned)
		if err := srv.Serve(l); err != nil && err != http.ErrServerClosed {
			t.Errorf("serve failed: %v", err)
		}
	}()

	requestErr := make(chan error, 1)
	go func() {
		resp, err := http.Get(fmt.Sprintf("http://%s/slow", l.Addr()))
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				err = fmt.Errorf("unexpected status %d", resp.StatusCode)
			}
		}
		requestErr <- err
	}()

	<-entered
	cancel()

	<-serveReturned
	wg.Wait()

	select {
	case <-requestDone:
	default:
		t.Fatal("shutdown wait released before the in-flight request finished")
	}
	if err := <-requestErr; err != nil {
		t.Fatalf("in-flight request failed: %v", err)
	}
}

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("could not find a free port: %v", err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func waitForHTTP(t *testing.T, url string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("server did not respond at %s in time", url)
}

func TestServeConnectReturnsAfterShutdownOnContextCancel(t *testing.T) {
	tests := []struct {
		name        string
		withMetrics bool
	}{
		{name: "connect server only", withMetrics: false},
		{name: "connect and metrics servers", withMetrics: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))

			var cfg Config
			cfg.Connect.Port = freePort(t)
			if tc.withMetrics {
				cfg.MetricsPort = freePort(t)
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			served := make(chan error, 1)
			go func() {
				served <- ServeConnect(ctx, logger, cfg, api.Deps{}, prometheus.NewRegistry())
			}()

			waitForHTTP(t, fmt.Sprintf("http://127.0.0.1:%d/ping", cfg.Connect.Port))
			if tc.withMetrics {
				waitForHTTP(t, fmt.Sprintf("http://127.0.0.1:%d/metrics", cfg.MetricsPort))
			}

			cancel()

			select {
			case err := <-served:
				if err != nil {
					t.Fatalf("ServeConnect returned error: %v", err)
				}
			case <-time.After(15 * time.Second):
				t.Fatal("ServeConnect did not return after context cancel")
			}
		})
	}
}
