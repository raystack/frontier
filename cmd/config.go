package cmd

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/raystack/salt/config"
	"github.com/spf13/cobra"
)

type Config struct {
	Host string `mapstructure:"host"`
}

func LoadConfig() (*Config, error) {
	var cfg Config

	loader := config.NewLoader(config.WithAppConfig("frontier"))
	err := loader.Load(&cfg)

	return &cfg, err
}

func configCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <command>",
		Short: "Manage client configurations",
		Example: heredoc.Doc(`
			$ frontier config init
			$ frontier config list`),
	}

	cmd.AddCommand(configInitCommand())
	cmd.AddCommand(configListCommand())

	return cmd
}

func configInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new client configuration",
		Example: heredoc.Doc(`
			$ frontier config init
		`),
		Annotations: map[string]string{
			"group": "core",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			loader := config.NewLoader(config.WithAppConfig("frontier"))

			if err := loader.Init(&Config{}); err != nil {
				return err
			}

			fmt.Println("config created")
			return nil
		},
	}
}

func configListCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "List client configuration settings",
		Example: heredoc.Doc(`
			$ frontier config list
		`),
		Annotations: map[string]string{
			"group": "core",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			loader := config.NewLoader(config.WithAppConfig("frontier"))

			data, err := loader.View()
			if err != nil {
				return ErrClientConfigNotFound
			}

			fmt.Println(data)
			return nil
		},
	}
	return cmd
}
