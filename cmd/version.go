package cmd

import (
	"fmt"

	"github.com/raystack/frontier/config"

	"github.com/raystack/salt/version"
	"github.com/spf13/cobra"
)

// VersionCmd prints the version of the binary
func versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			if config.Version == "" {
				fmt.Println("Version information not available")
				return nil
			}
			fmt.Println("Frontier: A secure and easy-to-use Authentication & Authorization Server")
			fmt.Printf("Version: %s\nBuild date: %s\nCommit: %s", config.Version, config.BuildDate, config.BuildCommit)
			fmt.Println(version.UpdateNotice(config.Version, "raystack/frontier"))
			return nil
		},
	}
}
