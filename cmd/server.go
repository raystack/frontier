package cmd

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/odpf/shield/config"
	"github.com/odpf/shield/internal/store/postgres/migrations"
	"github.com/odpf/shield/pkg/db"
	shieldlogger "github.com/odpf/shield/pkg/logger"
	"github.com/spf13/cobra"
	cli "github.com/spf13/cobra"
)

func ServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "server <command>",
		Aliases: []string{"s"},
		Short:   "Server management",
		Long:    "Server management commands.",
		Example: heredoc.Doc(`
			$ shield server start
			$ shield server start -c ./config.yaml
			$ shield server migrate
			$ shield server migrate -c ./config.yaml
			$ shield server migrate-rollback
			$ shield server migrate-rollback -c ./config.yaml
		`),
	}

	cmd.AddCommand(startCommand())
	cmd.AddCommand(migrateCommand())
	cmd.AddCommand(migrateRollbackCommand())

	return cmd
}

func startCommand() *cobra.Command {
	var configFile string

	c := &cli.Command{
		Use:     "start",
		Short:   "Start server and proxy default on port 8080",
		Example: "shield server start",
		RunE: func(cmd *cli.Command, args []string) error {
			appConfig, err := config.Load(configFile)
			if err != nil {
				panic(err)
			}
			logger := shieldlogger.InitLogger(appConfig.Log)

			return serve(logger, appConfig)
		},
	}

	c.Flags().StringVarP(&configFile, "config", "c", "./config.yaml", "Config file path")
	return c
}

func migrateCommand() *cobra.Command {
	var configFile string

	c := &cli.Command{
		Use:     "migrate",
		Short:   "Run DB Schema Migrations",
		Example: "shield migrate",
		RunE: func(c *cli.Command, args []string) error {
			appConfig, err := config.Load(configFile)
			if err != nil {
				panic(err)
			}

			return db.RunMigrations(db.Config{
				Driver: appConfig.DB.Driver,
				URL:    appConfig.DB.URL,
			}, migrations.MigrationFs, migrations.ResourcePath)
		},
	}

	c.Flags().StringVarP(&configFile, "config", "c", "./config.yaml", "Config file path")
	return c
}

func migrateRollbackCommand() *cobra.Command {
	var configFile string

	c := &cli.Command{
		Use:     "migration-rollback",
		Short:   "Run DB Schema Migrations Rollback to last state",
		Example: "shield migration-rollback",
		RunE: func(c *cli.Command, args []string) error {
			appConfig, err := config.Load(configFile)
			if err != nil {
				panic(err)
			}

			return db.RunRollback(db.Config{
				Driver: appConfig.DB.Driver,
				URL:    appConfig.DB.URL,
			}, migrations.MigrationFs, migrations.ResourcePath)
		},
	}

	c.Flags().StringVarP(&configFile, "config", "c", "./config.yaml", "Config file path")
	return c
}
