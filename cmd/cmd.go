package cmd

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/odpf/salt/cmdx"
	"github.com/spf13/cobra"
	cli "github.com/spf13/cobra"
)

func New(cfg *Config) *cli.Command {
	cliConfig = cfg

	var cmd = &cli.Command{
		Use:   "shield <command> <subcommand> [flags]",
		Short: "A cloud native role-based authorization aware reverse-proxy service",
		Long: heredoc.Doc(`
		A cloud native role-based authorization aware reverse-proxy service.`),
		SilenceUsage:  true,
		SilenceErrors: true,
		Example: heredoc.Doc(`
			$ shield group list
			$ shield organization list
			$ shield project list
			$ shield user create --file user.yaml
		`),
		Annotations: map[string]string{
			"group:core": "true",
			"help:learn": heredoc.Doc(`
				Use 'shield <command> <subcommand> --help' for more information about a command.
				Read the manual at https://odpf.github.io/shield/
			`),
			"help:feedback": heredoc.Doc(`
				Open an issue here https://github.com/odpf/shield/issues
			`),
			"help:environment": heredoc.Doc(`
				See 'shield help environment' for the list of supported environment variables.
			`),
		},
	}

	cmd.PersistentPreRunE = func(subCmd *cobra.Command, args []string) error {
		if IsClientCLI(subCmd) {
			if !IsClientConfigHostExist(subCmd) {
				return ErrClientConfigHostNotFound
			}
		}
		return nil
	}

	cmd.AddCommand(ServerCommand())
	cmd.AddCommand(NamespaceCommand(cliConfig))
	cmd.AddCommand(UserCommand(cliConfig))
	cmd.AddCommand(OrganizationCommand(cliConfig))
	cmd.AddCommand(GroupCommand(cliConfig))
	cmd.AddCommand(ProjectCommand(cliConfig))
	cmd.AddCommand(RoleCommand(cliConfig))
	cmd.AddCommand(ActionCommand(cliConfig))
	cmd.AddCommand(PolicyCommand(cliConfig))
	cmd.AddCommand(configCommand())

	// Help topics
	cmdx.SetHelp(cmd)
	cmd.AddCommand(cmdx.SetCompletionCmd("shield"))
	cmd.AddCommand(cmdx.SetHelpTopic("environment", envHelp))
	cmd.AddCommand(cmdx.SetHelpTopic("auth", authHelp))
	cmd.AddCommand(cmdx.SetRefCmd(cmd))
	return cmd
}
