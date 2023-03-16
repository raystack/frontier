package testbench

import (
	"log"

	"github.com/goto/shield/cmd"
	"github.com/goto/shield/config"
	"github.com/goto/shield/internal/store/postgres/migrations"
	"github.com/goto/shield/pkg/db"
	shieldlogger "github.com/goto/shield/pkg/logger"
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
