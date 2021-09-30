package cmd

import (
	"go.uber.org/zap"

	"google.golang.org/grpc"

	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpcctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	newrelic "github.com/newrelic/go-agent"
	"github.com/newrelic/go-agent/_integrations/nrgrpc"
	"github.com/odpf/salt/log"
	"github.com/odpf/shield/config"
)

func setupNewRelic(cfg config.NewRelic, logger log.Logger) newrelic.Application {
	nrCfg := newrelic.NewConfig(cfg.AppName, cfg.License)
	nrCfg.Enabled = cfg.Enabled

	nrApp, err := newrelic.NewApplication(nrCfg)
	if err != nil {
		logger.Fatal("failed to load Newrelic Application")
	}
	return nrApp
}

// REVISIT: passing config.Shield as reference
func getGRPCMiddleware(cfg *config.Shield, logger log.Logger) grpc.ServerOption {
	return grpc.UnaryInterceptor(
		grpcMiddleware.ChainUnaryServer(
			grpcRecovery.UnaryServerInterceptor(),
			grpcctxtags.UnaryServerInterceptor(),
			grpczap.UnaryServerInterceptor(zap.NewExample()),
			nrgrpc.UnaryServerInterceptor(setupNewRelic(cfg.NewRelic, logger)),
		))
}
