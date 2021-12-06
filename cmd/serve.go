package cmd

import (
	"context"
	"fmt"
	"github.com/odpf/salt/log"
	"github.com/odpf/salt/server"
	"github.com/odpf/shield/api/handler"
	v1 "github.com/odpf/shield/api/handler/v1"
	"github.com/odpf/shield/config"
	"github.com/odpf/shield/internal/group"
	"github.com/odpf/shield/internal/org"
	"github.com/odpf/shield/internal/project"
	"github.com/odpf/shield/internal/roles"
	"github.com/odpf/shield/internal/user"
	"github.com/odpf/shield/pkg/sql"
	"github.com/odpf/shield/proxy"
	blobstore "github.com/odpf/shield/store/blob"
	"github.com/odpf/shield/store/postgres"
	"github.com/pkg/errors"
	"github.com/pkg/profile"
	cli "github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	proxyTermChan         = make(chan os.Signal, 1)
	ruleCacheRefreshDelay = time.Minute * 2
)

func serveCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	c := &cli.Command{
		Use:     "serve",
		Short:   "Start server and proxy default on port 8080",
		Example: "shield serve",
		RunE: func(cmd *cli.Command, args []string) error {
			return serve(logger, appConfig)
		},
	}
	return c
}

func serve(logger log.Logger, appConfig *config.Shield) error {

	if profiling := os.Getenv("SHIELD_PROFILE"); profiling == "true" || profiling == "1" {
		defer profile.Start(profile.CPUProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()
	}

	ctx, cancelFunc := context.WithCancel(server.HandleSignals(context.Background()))
	defer cancelFunc()

	// set up shield proxy - h2c reverse proxy
	var cleanUpFunc []func() error
	var cleanUpProxies []func(ctx context.Context) error
	for _, service := range appConfig.Proxy.Services {
		if service.RulesPath == "" {
			return errors.New("ruleset field cannot be left empty")
		}
		blobFS, err := (&blobFactory{}).New(ctx, service.RulesPath, service.RulesPathSecret)
		if err != nil {
			return err
		}

		// TODO: option to use default http round tripper for http1.1 backends
		h2cProxy := proxy.NewH2c(proxy.NewH2cRoundTripper(logger), proxy.NewDirector())

		ruleRepo := blobstore.NewRuleRepository(logger, blobFS)
		if err := ruleRepo.InitCache(ctx, ruleCacheRefreshDelay); err != nil {
			return err
		}
		cleanUpFunc = append(cleanUpFunc, ruleRepo.Close)
		pipeline := buildPipeline(logger, h2cProxy, ruleRepo)
		go func(thisService config.Service, handler http.Handler) {
			proxyURL := fmt.Sprintf("%s:%d", thisService.Host, thisService.Port)
			logger.Info("starting h2c proxy", "url", proxyURL)

			//s, err := server.NewMux(server.Config{
			//	Port: appConfig.App.Port,
			//}, server.WithMuxGRPCServerOptions(getGRPCMiddleware(appConfig, logger)), nil)
			//s.RegisterHandler("/ping", healthCheck())
			//s.RegisterHandler("/", handler)

			mux := http.NewServeMux()
			mux.Handle("/ping", healthCheck())
			mux.Handle("/", handler)

			//create a tcp listener
			proxyListener, err := net.Listen("tcp", proxyURL)
			if err != nil {
				logger.Fatal("failed to listen", "err", err)
			}

			proxySrv := http.Server{
				Addr:    proxyURL,
				Handler: h2c.NewHandler(mux, &http2.Server{}),
			}
			if err := proxySrv.Serve(proxyListener); err != nil && err != http.ErrServerClosed {
				logger.Fatal("failed to serve", "err", err)
			}
			cleanUpProxies = append(cleanUpProxies, proxySrv.Shutdown)
		}(service, pipeline)
	}
	time.Sleep(100 * time.Millisecond)
	logger.Info("[shield] proxy is up")

	// set up shield api

	db, dbShutdown := setupDB(appConfig.DB, logger)
	defer dbShutdown()

	s, err := server.NewMux(server.Config{
		Port: appConfig.App.Port,
	}, server.WithMuxGRPCServerOptions(getGRPCMiddleware(appConfig, logger)))
	if err != nil {
		panic(err)
	}

	gw, err := server.NewGateway("", appConfig.App.Port)
	if err != nil {
		panic(err)
	}

	handler.Register(ctx, s, gw, apiDependencies(db, appConfig))

	go s.Serve()

	logger.Info("[shield] api is up")

	// we'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	signal.Notify(proxyTermChan, os.Interrupt, os.Kill, syscall.SIGTERM)

	// block until we receive our signal
	waitForTermSignal(ctx)

	// cleanup proxy stuff before shutdown
	for _, f := range cleanUpFunc {
		if err := f(); err != nil {
			logger.Warn("error occurred during shutdown", "err", err)
		}
	}
	for _, f := range cleanUpProxies {
		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, time.Second*20)
		if err := f(shutdownCtx); err != nil {
			shutdownCancel()
			logger.Warn("error occurred during shutdown", "err", err)
			continue
		}
		shutdownCancel()
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*10)
	defer shutdownCancel()

	s.Shutdown(shutdownCtx)

	return nil
}

func waitForTermSignal(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("process: ctx done bye\n")
			return
		case <-proxyTermChan:
			fmt.Printf("process: kill signal received. bye \n")
			return
		case <-time.After(time.Second * 5):
		}
	}
}

func healthCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "pong")
	}
}

func apiDependencies(db *sql.SQL, appConfig *config.Shield) handler.Deps {
	serviceStore := postgres.NewStore(db)

	dependencies := handler.Deps{
		V1: v1.Dep{
			OrgService: org.Service{
				Store: serviceStore,
			},
			UserService: user.Service{
				Store: serviceStore,
			},
			ProjectService: project.Service{
				Store: serviceStore,
			},
			RoleService: roles.Service{
				Store: postgres.NewStore(db),
			},
			GroupService: group.Service{
				Store: serviceStore,
			},
			IdentityProxyHeader: appConfig.App.IdentityProxyHeader,
		},
	}
	return dependencies
}
