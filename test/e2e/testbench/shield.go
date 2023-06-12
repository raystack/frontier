package testbench

import (
	"github.com/raystack/salt/log"
	"github.com/raystack/shield/cmd"
	"github.com/raystack/shield/config"
)

func MigrateShield(logger *log.Zap, appConfig *config.Shield) error {
	return cmd.RunMigrations(logger, appConfig.DB)
}

func StartShield(logger *log.Zap, appConfig *config.Shield) {
	go func() {
		if err := cmd.StartServer(logger, appConfig); err != nil {
			logger.Fatal("err starting", "err", err)
			panic(err)
		}
	}()
}
