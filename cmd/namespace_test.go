package cmd_test

import (
	"bytes"
	"testing"

	"github.com/raystack/frontier/cmd"
	"github.com/stretchr/testify/assert"
)

func TestClientNamespace(t *testing.T) {
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
				name:        "`namespace` list only should throw error host not found",
				want:        "",
				subCommands: []string{"list"},
				err:         cmd.ErrClientConfigHostNotFound,
			},
			{
				name:        "`namespace` list with host flag should return error",
				want:        "",
				subCommands: []string{"list", "-h", "test"},
				wantErr:     true,
			},
			{
				name:        "`namespace` view without host should throw error host not found",
				want:        "",
				subCommands: []string{"view", "123"},
				err:         cmd.ErrClientConfigHostNotFound,
			},
			{
				name:        "`namespace` view with host flag should return error",
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
				args := append([]string{"namespace"}, tt.subCommands...)
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
