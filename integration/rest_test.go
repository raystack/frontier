package integration_test

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/odpf/salt/log"
	"github.com/odpf/shield/middleware/basic_auth"
	"github.com/odpf/shield/middleware/authz"
	"github.com/odpf/shield/middleware/prefix"
	"github.com/odpf/shield/middleware/rulematch"
	"github.com/odpf/shield/proxy"
	"github.com/odpf/shield/store"
	blobstore "github.com/odpf/shield/store/blob"
	"github.com/stretchr/testify/assert"
	"gocloud.dev/blob/fileblob"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const (
	restBackendPort = 13777
	restProxyPort   = restBackendPort + 1
)

func TestREST(t *testing.T) {
	baseCtx, baseCancel := context.WithCancel(context.Background())
	defer baseCancel()

	blobFS, err := fileblob.OpenBucket("./fixtures", &fileblob.Options{
		CreateDir: true,
		Metadata:  fileblob.MetadataDontWrite,
	})
	if err != nil {
		t.Fatal(err)
	}

	h2cProxy := proxy.NewH2c(proxy.NewH2cRoundTripper(log.NewNoop()), proxy.NewDirector())
	ruleRepo := blobstore.NewRuleRepository(log.NewNoop(), blobFS)
	if err := ruleRepo.InitCache(baseCtx, time.Minute); err != nil {
		t.Fatal(err)
	}
	defer ruleRepo.Close()
	pipeline := buildPipeline(log.NewNoop(), h2cProxy, ruleRepo)

	proxyURL := fmt.Sprintf(":%d", restProxyPort)
	mux := http.NewServeMux()
	mux.Handle("/", pipeline)

	//create a tcp listener
	proxyListener, err := net.Listen("tcp", proxyURL)
	if err != nil {
		t.Fatal(err)
	}
	proxySrv := http.Server{
		Addr:    proxyURL,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}
	defer proxySrv.Close()
	go func() {
		if err := proxySrv.Serve(proxyListener); err != nil && err != http.ErrServerClosed {
			t.Error(err)
		}
	}()

	ts := startTestHTTPServer(restBackendPort, http.StatusOK, "")
	defer ts.Close()

	// wait for proxy to start
	time.Sleep(time.Second * 1)

	t.Run("should handle GET request with 200", func(t *testing.T) {
		backendReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/basic-authn/", restProxyPort), nil)
		if err != nil {
			assert.Nil(t, err)
		}
		backendReq.SetBasicAuth("user", "password")
		resp, err := http.DefaultClient.Do(backendReq)
		if err != nil {
			assert.Nil(t, err)
		}
		assert.Equal(t, 200, resp.StatusCode)
		resp.Body.Close()
	})
	t.Run("should give 401 if authn fails", func(t *testing.T) {
		backendReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/basic-authn/", restProxyPort), nil)
		if err != nil {
			assert.Nil(t, err)
		}
		backendReq.SetBasicAuth("user", "XX")
		resp, err := http.DefaultClient.Do(backendReq)
		if err != nil {
			assert.Nil(t, err)
		}
		assert.Equal(t, 401, resp.StatusCode)
		resp.Body.Close()
	})
	t.Run("should give 401 if authz fails on json payload", func(t *testing.T) {
		buff := bytes.NewReader([]byte(`{"project": "xx"}`))
		backendReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://127.0.0.1:%d/basic-authz/", restProxyPort), buff)
		if err != nil {
			t.Fatal(err)
		}
		backendReq.SetBasicAuth("user", "password")
		resp, err := http.DefaultClient.Do(backendReq)
		if err != nil {
			assert.Nil(t, err)
		}
		assert.Equal(t, 401, resp.StatusCode)
		resp.Body.Close()
	})
	t.Run("should give 200 if authz success on json payload", func(t *testing.T) {
		buff := bytes.NewReader([]byte(`{"project": "foo"}`))
		backendReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://127.0.0.1:%d/basic-authz/", restProxyPort), buff)
		if err != nil {
			t.Fatal(err)
		}
		backendReq.SetBasicAuth("user", "password")
		resp, err := http.DefaultClient.Do(backendReq)
		if err != nil {
			assert.Nil(t, err)
		}
		assert.Equal(t, 200, resp.StatusCode)
		resp.Body.Close()
	})
}

func BenchmarkProxyOverHttp1(b *testing.B) {
	baseCtx, baseCancel := context.WithCancel(context.Background())
	defer baseCancel()

	blobFS, err := fileblob.OpenBucket("./fixtures", &fileblob.Options{
		CreateDir: true,
		Metadata:  fileblob.MetadataDontWrite,
	})
	if err != nil {
		b.Fatal(err)
	}

	h2cProxy := proxy.NewH2c(proxy.NewH2cRoundTripper(log.NewNoop()), proxy.NewDirector())
	ruleRepo := blobstore.NewRuleRepository(log.NewNoop(), blobFS)
	if err := ruleRepo.InitCache(baseCtx, time.Minute); err != nil {
		b.Fatal(err)
	}
	defer ruleRepo.Close()
	pipeline := buildPipeline(log.NewNoop(), h2cProxy, ruleRepo)

	proxyURL := fmt.Sprintf(":%d", restProxyPort)
	mux := http.NewServeMux()
	mux.Handle("/", pipeline)

	//create a tcp listener
	proxyListener, err := net.Listen("tcp", proxyURL)
	if err != nil {
		b.Fatal(err)
	}
	proxySrv := http.Server{
		Addr:    proxyURL,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}
	defer proxySrv.Close()
	go func() {
		if err := proxySrv.Serve(proxyListener); err != nil && err != http.ErrServerClosed {
			b.Error(err)
		}
	}()

	ts := startTestHTTPServer(restBackendPort, http.StatusOK, "")
	defer ts.Close()

	// wait for proxy to start
	time.Sleep(time.Second * 1)

	b.Run("200 status code GET on http1.1", func(b *testing.B) {
		backendReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/basic/", restProxyPort), nil)
		if err != nil {
			b.Fatal(err)
		}
		for i := 0; i < b.N; i++ {
			resp, err := http.DefaultClient.Do(backendReq)
			if err != nil {
				panic(err)
			}
			if 200 != resp.StatusCode {
				b.Fatal("response code non 200")
			}
			resp.Body.Close()
		}
	})
	b.Run("200 status code with basic md5 authn on http1.1", func(b *testing.B) {
		backendReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/basic-authn/", restProxyPort), nil)
		if err != nil {
			b.Fatal(err)
		}
		backendReq.SetBasicAuth("user", "password")
		for i := 0; i < b.N; i++ {
			resp, err := http.DefaultClient.Do(backendReq)
			if err != nil {
				panic(err)
			}
			if 200 != resp.StatusCode {
				b.Fatal("response code non 200")
			}
			resp.Body.Close()
		}
	})
	b.Run("200 status code with basic bcrypt authn on http1.1", func(b *testing.B) {
		backendReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/basic-authn-bcrypt/", restProxyPort), nil)
		if err != nil {
			b.Fatal(err)
		}
		backendReq.SetBasicAuth("user", "password")
		for i := 0; i < b.N; i++ {
			resp, err := http.DefaultClient.Do(backendReq)
			if err != nil {
				panic(err)
			}
			if 200 != resp.StatusCode {
				b.Fatal("response code non 200")
			}
			resp.Body.Close()
		}
	})
	b.Run("200 status code with basic authz on http1.1", func(b *testing.B) {
		buff := bytes.NewReader([]byte(`{"project": "foo"}`))
		backendReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://127.0.0.1:%d/basic-authz/", restProxyPort), buff)
		if err != nil {
			b.Fatal(err)
		}
		backendReq.SetBasicAuth("user", "password")
		for i := 0; i < b.N; i++ {
			resp, err := http.DefaultClient.Do(backendReq)
			if err != nil {
				panic(err)
			}
			if 200 != resp.StatusCode {
				b.Fatal("response code non 200")
			}
			resp.Body.Close()
		}
	})
}

// buildPipeline builds middleware sequence
func buildPipeline(logger log.Logger, proxy http.Handler, ruleRepo store.RuleRepository) http.Handler {
	// Note: execution order is bottom up
	prefixWare := prefix.New(logger, proxy)
	casbinAuthz := authz.New(logger, prefixWare)
	basicAuthn := basic_auth.New(logger, casbinAuthz)
	matchWare := rulematch.New(logger, basicAuthn, rulematch.NewRegexMatcher(ruleRepo))
	return matchWare
}

func startTestHTTPServer(port, statusCode int, content string) (ts *httptest.Server) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		if content != "" {
			_, err := w.Write([]byte(content))
			if err != nil {
				panic(err)
			}
		}
	})
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		panic(err)
	}
	ts = &httptest.Server{
		Listener: listener,
		Config: &http.Server{
			Handler:      h2c.NewHandler(handler, &http2.Server{}),
			ReadTimeout:  time.Second,
			WriteTimeout: time.Second,
			IdleTimeout:  time.Second,
		},
		EnableHTTP2: true,
	}
	ts.Start()
	return ts
}
