package sql

import (
	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(config Config, path ...string) error {
	m, err := getMigrationInstance(config, path)
	if err != nil {
		return err
	}

	err = m.Up()
	if err == migrate.ErrNoChange || err == nil {
		return nil
	}

	return err
}

func RunRollback(config Config, path ...string) error {
	m, err := getMigrationInstance(config, path)
	if err != nil {
		return err
	}

	err = m.Steps(-1)
	if err == migrate.ErrNoChange || err == nil {
		return nil
	}

	return err
}

func getMigrationInstance(config Config, path []string) (*migrate.Migrate, error) {
	db, err := sql.Open(config.Driver, config.URL)
	if err != nil {
		return nil, err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, err
	}

	migrationPath := "file://./migrations"
	if len(path) >= 1 && path[0] != "" {
		migrationPath = path[0]
	}

	m, err := migrate.NewWithDatabaseInstance(migrationPath, config.Driver, driver)
	if err != nil {
		return nil, err
	}

	return m, nil
}
