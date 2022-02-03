package cmd

import (
	"github.com/odpf/salt/log"
	"github.com/odpf/shield/config"
	cli "github.com/spf13/cobra"
)

func New(logger log.Logger, appConfig *config.Shield) *cli.Command {
	var cmd = &cli.Command{
		Use:          "shield",
		SilenceUsage: true,
	}

	cmd.AddCommand(serveCommand(logger, appConfig))
	cmd.AddCommand(migrationsCommand(logger, appConfig))
	cmd.AddCommand(migrationsRollbackCommand(logger, appConfig))
	cmd.AddCommand(NamespaceCommand(logger, appConfig))
	cmd.AddCommand(UserCommand(logger, appConfig))
	cmd.AddCommand(OrganizationCommand(logger, appConfig))
	cmd.AddCommand(GroupCommand(logger, appConfig))
	cmd.AddCommand(ProjectCommand(logger, appConfig))
	return cmd
}
