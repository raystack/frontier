package cmd

import (
	"context"
	"time"

	"github.com/odpf/salt/log"
	"github.com/odpf/salt/server"
	"github.com/odpf/shield/api/handler"
	v1 "github.com/odpf/shield/api/handler/v1"
	"github.com/odpf/shield/config"
	"github.com/odpf/shield/org"
	"github.com/odpf/shield/postgres"
	cli "github.com/spf13/cobra"
)

func apiCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	c := &cli.Command{
		Use:     "api",
		Short:   "Start shield api server",
		Example: "shield serve api",
		RunE: func(c *cli.Command, args []string) error {
			ctx, cancelFunc := context.WithCancel(server.HandleSignals(context.Background()))
			defer cancelFunc()

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

			handler.Register(ctx, s, gw, handler.Deps{
				V1: v1.Dep{
					OrgService: org.Service{
						Store: postgres.NewStore(db),
					},
				},
			})

			go s.Serve()
			<-ctx.Done()

			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*10)
			defer shutdownCancel()

			s.Shutdown(shutdownCtx)

			return nil
		},
	}
	return c
}
