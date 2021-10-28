package sql

import (
	"embed"
	"fmt"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
)

func RunMigrations(config Config, embeddedMigrations embed.FS, resourcePath string) error {
	m, err := getMigrationInstance(config, embeddedMigrations, resourcePath)
	if err != nil {
		return err
	}

	err = m.Up()
	if err == migrate.ErrNoChange || err == nil {
		return nil
	}

	return err
}

func RunRollback(config Config, embeddedMigrations embed.FS, resourcePath string) error {
	m, err := getMigrationInstance(config, embeddedMigrations, resourcePath)
	if err != nil {
		return err
	}

	err = m.Steps(-1)
	if err == migrate.ErrNoChange || err == nil {
		return nil
	}

	return err
}

func getMigrationInstance(config Config, embeddedMigrations embed.FS, resourcePath string) (*migrate.Migrate, error) {
	src, err := httpfs.New(http.FS(embeddedMigrations), resourcePath)
	if err != nil {
		return &migrate.Migrate{}, fmt.Errorf("db migrator: %v", err)
	}
	return migrate.NewWithSourceInstance("httpfs", src, config.URL)
}
