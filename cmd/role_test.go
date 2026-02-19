package cmd_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/raystack/frontier/cmd"
	"github.com/stretchr/testify/assert"
)

func TestClientRole(t *testing.T) {
	t.Run("without config file", func(t *testing.T) {
		tests := []struct {
			name        string
			cliConfig   *cmd.Config
			subCommands []string
			want        string
			err         error
			wantErr     bool
		}{
			{
				name:        "`role` list only should throw error host not found",
				want:        "",
				subCommands: []string{"list"},
				err:         cmd.ErrClientConfigHostNotFound,
			},
			{
				name:        "`role` list with host flag should return error",
				want:        "",
				subCommands: []string{"list", "-h", "test"},
				wantErr:     true,
			},
			{
				name:        "`role` create only should throw error host not found",
				want:        "",
				subCommands: []string{"create"},
				err:         cmd.ErrClientConfigHostNotFound,
			},
			{
				name:        "`role` create with host flag should throw error missing required flag",
				want:        "",
				subCommands: []string{"create", "-h", "test"},
				err:         errors.New("required flag(s) \"file\", \"header\" not set"),
			},
			{
				name:        "`role` edit without host should throw error host not found",
				want:        "",
				subCommands: []string{"edit", "123"},
				err:         cmd.ErrClientConfigHostNotFound,
			},
			{
				name:        "`role` edit with host flag should throw error missing required flag",
				want:        "",
				subCommands: []string{"edit", "123", "-h", "test"},
				err:         errors.New("required flag(s) \"file\" not set"),
			},
			{
				name:        "`role` view without host should throw error host not found",
				want:        "",
				subCommands: []string{"view", "123"},
				err:         cmd.ErrClientConfigHostNotFound,
			},
			{
				name:        "`role` view with host flag should return error",
				want:        "",
				subCommands: []string{"view", "123", "-h", "test"},
				wantErr:     true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tt.cliConfig = &cmd.Config{}
				cli := cmd.New(tt.cliConfig)

				buf := new(bytes.Buffer)
				cli.SetOut(buf)
				args := append([]string{"role"}, tt.subCommands...)
				cli.SetArgs(args)

				err := cli.Execute()
				got := buf.String()

				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.Equal(t, tt.err, err)
				}
				assert.Equal(t, tt.want, got)
			})
		}
	})
}
