package testbench

import (
	"github.com/odpf/salt/log"
	"github.com/odpf/shield/cmd"
	"github.com/odpf/shield/config"
	"github.com/odpf/shield/internal/store/postgres/migrations"
	"github.com/odpf/shield/pkg/db"
)

func MigrateShield(appConfig *config.Shield) error {
	return db.RunMigrations(db.Config{
		Driver: appConfig.DB.Driver,
		URL:    appConfig.DB.URL,
	}, migrations.MigrationFs, migrations.ResourcePath)
}

func StartShield(logger *log.Zap, appConfig *config.Shield) {
	go func() {
		if err := cmd.StartServer(logger, appConfig); err != nil {
			logger.Fatal("err starting", "err", err)
			panic(err)
		}
	}()
}
