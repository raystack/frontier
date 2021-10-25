package cmd

import (
	"github.com/odpf/salt/log"
	"github.com/odpf/shield/config"
	"github.com/odpf/shield/pkg/sql"
	"github.com/odpf/shield/store/postgres/migrations"
	cli "github.com/spf13/cobra"
)

func migrationsCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	c := &cli.Command{
		Use:     "migrate",
		Short:   "Run DB Schema Migrations",
		Example: "shield migrate",
		RunE: func(c *cli.Command, args []string) error {
			return sql.RunMigrations(sql.Config{
				Driver: appConfig.DB.Driver,
				URL:    appConfig.DB.URL,
			}, migrations.MigrationFs)
		},
	}
	return c
}

func migrationsRollbackCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	c := &cli.Command{
		Use:     "migration-rollback",
		Short:   "Run DB Schema Migrations Rollback to last state",
		Example: "shield migration-rollback",
		RunE: func(c *cli.Command, args []string) error {
			return sql.RunRollback(sql.Config{
				Driver: appConfig.DB.Driver,
				URL:    appConfig.DB.URL,
			}, migrations.MigrationFs)
		},
	}
	return c
}
