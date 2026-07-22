package cmd

import (
	"testing"

	cli "github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestMustMarkRequired(t *testing.T) {
	t.Run("should mark an existing flag as required", func(t *testing.T) {
		cmd := &cli.Command{Use: "test"}
		cmd.Flags().String("file", "", "")

		assert.NotPanics(t, func() { mustMarkRequired(cmd, "file") })
		assert.Error(t, cmd.ValidateRequiredFlags())
	})

	t.Run("should panic on an unknown flag name", func(t *testing.T) {
		cmd := &cli.Command{Use: "test"}

		assert.Panics(t, func() { mustMarkRequired(cmd, "no-such-flag") })
	})
}
