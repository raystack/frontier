package cmd_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/raystack/frontier/cmd"
	"github.com/stretchr/testify/assert"
)

func TestClientGroup(t *testing.T) {
	orgID := uuid.New().String()
	t.Run("without config file", func(t *testing.T) {
		tests := []struct {
			name        string
			cliConfig   *cmd.Config
			subCommands []string
			want        string
			err         error
		}{
			{
				name:        "`group` list only should throw error host not found",
				want:        "",
				subCommands: []string{"list", orgID},
				err:         cmd.ErrClientConfigHostNotFound,
			},
			{
				name:        "`group` list with host flag should pass",
				want:        "",
				subCommands: []string{"list", orgID, "-h", "test"},
				err:         context.DeadlineExceeded,
			},
			{
				name:        "`group` create only should throw error host not found",
				want:        "",
				subCommands: []string{"create"},
				err:         cmd.ErrClientConfigHostNotFound,
			},
			{
				name:        "`group` create with host flag should throw error missing required flag",
				want:        "",
				subCommands: []string{"create", "-h", "test"},
				err:         errors.New("required flag(s) \"file\", \"header\" not set"),
			},
			{
				name:        "`group` edit without host should throw error host not found",
				want:        "",
				subCommands: []string{"edit", "123"},
				err:         cmd.ErrClientConfigHostNotFound,
			},
			{
				name:        "`group` edit with host flag should throw error missing required flag",
				want:        "",
				subCommands: []string{"edit", "123", "-h", "test"},
				err:         errors.New("required flag(s) \"file\" not set"),
			},
			{
				name:        "`group` view without host should throw error host not found",
				want:        "",
				subCommands: []string{"view", orgID, "123"},
				err:         cmd.ErrClientConfigHostNotFound,
			},
			{
				name:        "`group` view with host flag should pass",
				want:        "",
				subCommands: []string{"view", orgID, "123", "-h", "test"},
				err:         context.DeadlineExceeded,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tt.cliConfig = &cmd.Config{}
				cli := cmd.New(tt.cliConfig)

				buf := new(bytes.Buffer)
				cli.SetOut(buf)
				args := append([]string{"group"}, tt.subCommands...)
				cli.SetArgs(args)

				err := cli.Execute()
				got := buf.String()

				assert.Equal(t, tt.err, err)
				assert.Equal(t, tt.want, got)
			})
		}
	})
}
