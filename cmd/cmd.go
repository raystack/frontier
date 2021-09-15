package cmd

import (
	cli "github.com/spf13/cobra"
)

func New() *cli.Command {
	var cmd = &cli.Command{
		Use:          "proxy",
		SilenceUsage: true,
	}

	cmd.AddCommand(serveCommand())

	return cmd
}