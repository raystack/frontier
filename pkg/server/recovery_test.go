package server

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestConnectPanicRecoveryReturnsInternalError(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))

	handler := connect.NewUnaryHandler(
		"/test.v1.TestService/Panic",
		func(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
			panic("boom")
		},
		connectPanicRecovery(logger),
	)
	mux := http.NewServeMux()
	mux.Handle("/test.v1.TestService/Panic", handler)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/test.v1.TestService/Panic", "application/json", strings.NewReader("{}"))
	if err != nil {
		t.Fatalf("expected an error response, got a dropped connection: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if !strings.Contains(string(body), `"internal"`) {
		t.Errorf("body %q does not carry the internal error code", body)
	}
	if strings.Contains(string(body), "boom") {
		t.Errorf("panic detail leaked to the client: %q", body)
	}

	logged := logBuf.String()
	if !strings.Contains(logged, `"level":"ERROR"`) {
		t.Errorf("panic was not logged at error level: %q", logged)
	}
	if !strings.Contains(logged, "boom") || !strings.Contains(logged, "/test.v1.TestService/Panic") {
		t.Errorf("log entry is missing the panic value or procedure: %q", logged)
	}
	if !strings.Contains(logged, "recovery_test.go") {
		t.Errorf("log entry is missing the stack trace: %q", logged)
	}
}

func TestUIPanicRecoveryReturnsInternalError(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))

	mux := http.NewServeMux()
	mux.HandleFunc("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	rec := httptest.NewRecorder()
	uiPanicRecovery(logger, mux).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/panic", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if strings.Contains(rec.Body.String(), "boom") {
		t.Errorf("panic detail leaked to the client: %q", rec.Body.String())
	}

	logged := logBuf.String()
	if !strings.Contains(logged, `"level":"ERROR"`) || !strings.Contains(logged, "boom") || !strings.Contains(logged, "/panic") {
		t.Errorf("log entry is missing the error level, panic value, or path: %q", logged)
	}
}

func TestUIPanicRecoveryPassesThroughNormalRequests(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	rec := httptest.NewRecorder()
	uiPanicRecovery(logger, mux).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ok", nil))

	if rec.Code != http.StatusTeapot {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusTeapot)
	}
}

func TestUIPanicRecoveryAbortsCommittedResponses(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))

	mux := http.NewServeMux()
	mux.HandleFunc("/partial", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("partial")); err != nil {
			t.Fatalf("write: %v", err)
		}
		panic("boom")
	})

	rec := httptest.NewRecorder()
	func() {
		defer func() {
			if recovered := recover(); recovered != http.ErrAbortHandler {
				t.Errorf("recovered %v, want http.ErrAbortHandler", recovered)
			}
		}()
		uiPanicRecovery(logger, mux).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/partial", nil))
	}()

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want the already-committed %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "partial" {
		t.Errorf("body = %q, want the partial body with nothing appended", rec.Body.String())
	}
	if !strings.Contains(logBuf.String(), "boom") {
		t.Errorf("panic was not logged: %q", logBuf.String())
	}
}

func TestUIPanicRecoveryAbortsAfterFlush(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	mux := http.NewServeMux()
	mux.HandleFunc("/flush", func(w http.ResponseWriter, r *http.Request) {
		w.(http.Flusher).Flush()
		panic("boom")
	})

	rec := httptest.NewRecorder()
	func() {
		defer func() {
			if recovered := recover(); recovered != http.ErrAbortHandler {
				t.Errorf("recovered %v, want http.ErrAbortHandler", recovered)
			}
		}()
		uiPanicRecovery(logger, mux).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/flush", nil))
	}()

	if strings.Contains(rec.Body.String(), "internal server error") {
		t.Errorf("error body appended to a flushed response: %q", rec.Body.String())
	}
}

func TestUIPanicRecoveryRepanicsOnErrAbortHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	mux := http.NewServeMux()
	mux.HandleFunc("/abort", func(w http.ResponseWriter, r *http.Request) {
		panic(http.ErrAbortHandler)
	})

	defer func() {
		if recovered := recover(); recovered != http.ErrAbortHandler {
			t.Errorf("recovered %v, want http.ErrAbortHandler", recovered)
		}
	}()
	uiPanicRecovery(logger, mux).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/abort", nil))
}
