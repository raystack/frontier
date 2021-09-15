package cmd

import (
	cli "github.com/spf13/cobra"
)

func serveCommand() *cli.Command {
	c := &cli.Command{
		Use:   "serve",
		Short: "Start server and proxy default on port 8080",
	}
	c.AddCommand(proxyCommand())
	c.AddCommand(adminCommand())
	return c
}