package testbench

import (
	"log"

	"github.com/raystack/shield/cmd"
	"github.com/raystack/shield/config"
	"github.com/raystack/shield/internal/store/postgres/migrations"
	"github.com/raystack/shield/pkg/db"
	shieldlogger "github.com/raystack/shield/pkg/logger"
)

func migrateShield(appConfig *config.Shield) error {
	return db.RunMigrations(db.Config{
		Driver: appConfig.DB.Driver,
		URL:    appConfig.DB.URL,
	}, migrations.MigrationFs, migrations.ResourcePath)
}

func startShield(appConfig *config.Shield) {
	go func() {
		logger := shieldlogger.InitLogger(appConfig.Log)
		if err := cmd.StartServer(logger, appConfig); err != nil {
			log.Fatal(err)
		}
	}()
}
