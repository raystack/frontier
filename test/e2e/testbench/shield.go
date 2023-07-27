package testbench

import (
	"github.com/raystack/frontier/cmd"
	"github.com/raystack/frontier/config"
	"github.com/raystack/salt/log"
)

func MigrateFrontier(logger *log.Zap, appConfig *config.Frontier) error {
	return cmd.RunMigrations(logger, appConfig.DB)
}

func StartFrontier(logger *log.Zap, appConfig *config.Frontier) {
	go func() {
		if err := cmd.StartServer(logger, appConfig); err != nil {
			logger.Fatal("err starting", "err", err)
			panic(err)
		}
	}()
}
