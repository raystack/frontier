package cmd

import (
	"github.com/odpf/salt/log"
	"github.com/odpf/shield/config"
	cli "github.com/spf13/cobra"
)

func New(logger log.Logger, appConfig *config.Shield) *cli.Command {
	var cmd = &cli.Command{
		Use:          "proxy",
		SilenceUsage: true,
	}

	cmd.AddCommand(serveCommand(logger, appConfig))
	cmd.AddCommand(migrationsCommand(logger, appConfig))
	cmd.AddCommand(migrationsRollbackCommand(logger, appConfig))
	return cmd
}
