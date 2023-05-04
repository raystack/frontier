package cmd

import (
	"fmt"

	"github.com/odpf/salt/version"
	"github.com/spf13/cobra"
)

// VersionCmd prints the version of the binary
func versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			if Version == "" {
				fmt.Println("Version information not available")
				return nil
			}

			fmt.Printf("shield version %s\t", Version)
			fmt.Println(version.UpdateNotice(Version, "odpf/shield"))
			return nil
		},
	}
}
