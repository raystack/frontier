package cmd

import (
	"github.com/odpf/salt/log"
	"github.com/odpf/shield/config"
	"github.com/odpf/shield/internal/store/postgres/migrations"
	"github.com/odpf/shield/pkg/db"
	cli "github.com/spf13/cobra"
)

func migrationsCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	c := &cli.Command{
		Use:     "migrate",
		Short:   "Run DB Schema Migrations",
		Example: "shield migrate",
		RunE: func(c *cli.Command, args []string) error {
			return db.RunMigrations(db.Config{
				Driver: appConfig.DB.Driver,
				URL:    appConfig.DB.URL,
			}, migrations.MigrationFs, migrations.ResourcePath)
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
			return db.RunRollback(db.Config{
				Driver: appConfig.DB.Driver,
				URL:    appConfig.DB.URL,
			}, migrations.MigrationFs, migrations.ResourcePath)
		},
	}
	return c
}
