package cmd

import (
	"context"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/pkg/errors"
	"github.com/raystack/salt/log"
	"github.com/raystack/shield/core/metaschema"
	"github.com/raystack/shield/internal/store/postgres"
	"github.com/raystack/shield/internal/store/postgres/migrations"
	"github.com/raystack/shield/pkg/db"
)

func RunMigrations(logger log.Logger, config db.Config) error {
	m, err := getDatabaseMigrationInstance(config)
	if err != nil {
		return err
	}
	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	// populate default metaschemas in the database
	logger.Info("migrating default metadata schemas to db")
	dbc, err := db.New(config)
	if err != nil {
		return errors.Wrap(err, "failed to connect to db")
	}
	metaschemaRepository := postgres.NewMetaSchemaRepository(logger, dbc)
	metaschemaService := metaschema.NewService(metaschemaRepository)
	if err = metaschemaService.MigrateDefault(context.Background()); err != nil {
		return errors.Wrap(err, "failed to add default schemas to db")
	}

	return nil
}

func RunRollback(config db.Config) error {
	m, err := getDatabaseMigrationInstance(config)
	if err != nil {
		return err
	}

	err = m.Steps(-1)
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return err
}

func getDatabaseMigrationInstance(config db.Config) (*migrate.Migrate, error) {
	fs := migrations.MigrationFs
	resourcePath := migrations.ResourcePath
	src, err := iofs.New(fs, resourcePath)
	if err != nil {
		return &migrate.Migrate{}, fmt.Errorf("db migrator: %v", err)
	}
	return migrate.NewWithSourceInstance("iofs", src, config.URL)
}
