package cmd_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/odpf/shield/cmd"
	"github.com/stretchr/testify/assert"
)

var expectedOrganizationUsageHelp = heredoc.Doc(`

Usage
  shield organization [flags]

Core commands
  admadd      add admins to an organization
  admlist     list admins of an organization
  admremove   remove admins from an organization
  create      Create an organization
  edit        Edit an organization
  list        List all organizations
  view        View an organization

Flags
  -h, --host string   Shield API service to connect to

Inherited flags
  --help   Show help for command

Examples
  $ shield organization create
  $ shield organization edit
  $ shield organization view
  $ shield organization list

`)

func TestClientOrganization(t *testing.T) {
	t.Run("without config file", func(t *testing.T) {
		tests := []struct {
			name        string
			cliConfig   *cmd.Config
			subCommands []string
			want        string
			err         error
		}{
			{
				name: "`organization` only should show usage help",
				want: expectedOrganizationUsageHelp,
				err:  nil,
			},
			{
				name:        "`organization` list only should throw error host not found",
				want:        "",
				subCommands: []string{"list"},
				err:         cmd.ErrClientConfigHostNotFound,
			},
			{
				name:        "`organization` list with host flag should pass",
				want:        "",
				subCommands: []string{"list", "-h", "test"},
				err:         context.DeadlineExceeded,
			},
			{
				name:        "`organization` create only should throw error host not found",
				want:        "",
				subCommands: []string{"create"},
				err:         cmd.ErrClientConfigHostNotFound,
			},
			{
				name:        "`organization` create with host flag should throw error missing required flag",
				want:        "",
				subCommands: []string{"create", "-h", "test"},
				err:         errors.New("required flag(s) \"file\", \"header\" not set"),
			},
			{
				name:        "`organization` edit without host should throw error host not found",
				want:        "",
				subCommands: []string{"edit", "123"},
				err:         cmd.ErrClientConfigHostNotFound,
			},
			{
				name:        "`organization` edit with host flag should throw error missing required flag",
				want:        "",
				subCommands: []string{"edit", "123", "-h", "test"},
				err:         errors.New("required flag(s) \"file\" not set"),
			},
			{
				name:        "`organization` view without host should throw error host not found",
				want:        "",
				subCommands: []string{"view", "123"},
				err:         cmd.ErrClientConfigHostNotFound,
			},
			{
				name:        "`organization` view with host flag should pass",
				want:        "",
				subCommands: []string{"view", "123", "-h", "test"},
				err:         context.DeadlineExceeded,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tt.cliConfig = &cmd.Config{}
				cli := cmd.New(tt.cliConfig)

				buf := new(bytes.Buffer)
				cli.SetOutput(buf)
				args := append([]string{"organization"}, tt.subCommands...)
				cli.SetArgs(args)

				err := cli.Execute()
				got := buf.String()

				assert.Equal(t, tt.err, err)
				assert.Equal(t, tt.want, got)
			})
		}
	})
}
