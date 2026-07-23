package cmd

import (
	"fmt"

	cli "github.com/spf13/cobra"
)

// mustMarkRequired marks each named flag as required. It panics when a flag
// is not defined on the command, so a mistyped or renamed flag name fails on
// the first run instead of silently making the flag optional.
func mustMarkRequired(cmd *cli.Command, names ...string) {
	for _, name := range names {
		if err := cmd.MarkFlagRequired(name); err != nil {
			panic(fmt.Sprintf("marking --%s required on %q: %v", name, cmd.Name(), err))
		}
	}
}
