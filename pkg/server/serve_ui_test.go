package server

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestServeUIReturnsAfterShutdownOnContextCancel(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("could not find a free port: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		ServeUI(ctx, logger, UIConfig{Port: port}, Config{})
	}()

	url := fmt.Sprintf("http://127.0.0.1:%d/configs", port)
	deadline := time.Now().Add(10 * time.Second)
	for {
		if time.Now().After(deadline) {
			t.Fatalf("ui server did not respond at %s in time", url)
		}
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
	}

	cancel()

	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatal("ServeUI did not return after context cancel")
	}
}

func TestServeUIReturnsWhenPortNotConfigured(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	done := make(chan struct{})
	go func() {
		defer close(done)
		ServeUI(context.Background(), logger, UIConfig{}, Config{})
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("ServeUI did not return when no port is configured")
	}
}
