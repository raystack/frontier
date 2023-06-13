package proxy

import (
	"context"
	"fmt"
	"net/http"

	"github.com/raystack/salt/log"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func Serve(
	ctx context.Context,
	logger log.Logger,
	cfg Config,
	handler http.Handler,
) func(ctx context.Context) error {
	proxyURL := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	logger.Info("starting h2c proxy", "url", proxyURL)

	mux := http.NewServeMux()
	mux.Handle("/ping", healthCheck())
	mux.Handle("/", handler)

	proxySrv := http.Server{
		Addr:    proxyURL,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	go func(ctx context.Context, logger log.Logger, cfg Config) {
		if err := proxySrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to serve", "err", err)
		}

		logger.Info("[shield] proxy stopped", "service", cfg.Name)
	}(ctx, logger, cfg)

	logger.Info("[shield] proxy ready", "service", cfg.Name)
	return proxySrv.Shutdown
}

func healthCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "pong")
	}
}
