package cmd

import (
	"github.com/odpf/salt/log"
	"github.com/odpf/shield/config"
	cli "github.com/spf13/cobra"
)

func serveCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	c := &cli.Command{
		Use:   "serve",
		Short: "Start server and proxy default on port 8080",
	}
	c.AddCommand(proxyCommand(logger, appConfig))
	c.AddCommand(adminCommand())
	c.AddCommand(apiCommand())
	return c
}
