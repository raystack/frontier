package cmd

import (
	"context"
	"fmt"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/odpf/salt/server"
	"github.com/odpf/shield/echo"
	"github.com/odpf/shield/gen/go/protos"
	cli "github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net/http"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
)

var GRPCMiddlewaresInterceptor = grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
	grpc_recovery.UnaryServerInterceptor(),
	grpc_ctxtags.UnaryServerInterceptor(),
	grpc_zap.UnaryServerInterceptor(zap.NewExample()),
	//nrgrpc.UnaryServerInterceptor(app),
))

func apiCommand() *cli.Command {
	c := &cli.Command{
		Use:     "api",
		Short:   "Start serving admin endpoint",
		Example: "shield serve admin",
		RunE: func(c *cli.Command, args []string) error {
			ctx, cancelFunc := context.WithCancel(server.HandleSignals(context.Background()))
			defer cancelFunc()

			s, err := server.NewMux(server.Config{
				Port: 8000,
			}, server.WithMuxGRPCServerOptions(GRPCMiddlewaresInterceptor))
			if err != nil {
				panic(err)
			}

			gatewayClientPort := 8000
			gw, err := server.NewGateway("", gatewayClientPort)
			if err != nil {
				panic(err)
			}
			//gw.RegisterHandler(ctx, commonv1.RegisterCommonServiceHandlerFromEndpoint)
			gw.RegisterHandler(ctx, protos.RegisterYourServiceHandlerFromEndpoint)

			s.SetGateway("/api", gw)

			s.RegisterHandler("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, "pong")
			}))

			s.RegisterService(&protos.YourService_ServiceDesc,
				&echo.SimpleServer{},
			)

			go s.Serve()
			<-ctx.Done()
			// clean anything that needs to be closed etc like common server implementation etc
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*10)
			defer shutdownCancel()

			s.Shutdown(shutdownCtx)

			return nil
		},
	}
	return c
}
