package cmd

import (
	"fmt"

	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpcctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/newrelic/go-agent/v3/integrations/nrgrpc"
	newrelic "github.com/newrelic/go-agent/v3/newrelic"
	"github.com/odpf/salt/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.uber.org/zap"

	"github.com/odpf/shield/config"
	"github.com/odpf/shield/grpc_interceptors"
	"github.com/odpf/shield/pkg/sql"
)

func setupNewRelic(cfg config.NewRelic, logger log.Logger) *newrelic.Application {
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(cfg.AppName),
		newrelic.ConfigLicense(cfg.License),
		newrelic.ConfigEnabled(cfg.Enabled),
	)

	if err != nil {
		logger.Fatal("failed to load Newrelic Application")
	}

	return app
}

// REVISIT: passing config.Shield as reference
func getGRPCMiddleware(cfg *config.Shield, logger log.Logger) grpc.ServerOption {
	customFunc := func(p interface{}) (err error) {
		return status.Errorf(codes.Internal, "internal server error")
	}
	opts := []grpcRecovery.Option{
		grpcRecovery.WithRecoveryHandler(customFunc),
	}

	return grpc.UnaryInterceptor(
		grpcMiddleware.ChainUnaryServer(
			grpc_interceptors.EnrichCtxWithIdentity(cfg.App.IdentityProxyHeader),
			grpczap.UnaryServerInterceptor(zap.NewExample()),
			grpcRecovery.UnaryServerInterceptor(opts...),
			grpcctxtags.UnaryServerInterceptor(),
			nrgrpc.UnaryServerInterceptor(setupNewRelic(cfg.NewRelic, logger)),
		))
}

func setupDB(cfg config.DBConfig, logger log.Logger) (*sql.SQL, func()) {
	db, err := sql.New(sql.Config{
		Driver:              cfg.Driver,
		URL:                 cfg.URL,
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxOpenConns:        cfg.MaxOpenConns,
		ConnMaxLifeTime:     cfg.ConnMaxLifeTime,
		MaxQueryTimeoutInMS: cfg.MaxQueryTimeout,
	})

	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to setup db: %s", err.Error()))
	}

	return db, func() { db.Close() }
}
