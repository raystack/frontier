package cmd

import (
	"context"
	"github.com/odpf/shield/api/handler"
	v1 "github.com/odpf/shield/api/handler/v1"
	"github.com/odpf/shield/config"
	"github.com/odpf/shield/internal/group"
	"github.com/odpf/shield/internal/org"
	"github.com/odpf/shield/internal/project"
	"github.com/odpf/shield/internal/roles"
	"github.com/odpf/shield/internal/user"
	"github.com/odpf/shield/store/postgres"
	"time"

	"github.com/odpf/salt/log"
	"github.com/odpf/salt/server"
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

			serviceStore := postgres.NewStore(db)
			handler.Register(ctx, s, gw, handler.Deps{
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
