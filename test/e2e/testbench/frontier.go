package testbench

import (
	"log/slog"

	"github.com/raystack/frontier/cmd"
	"github.com/raystack/frontier/config"
	frontierlogger "github.com/raystack/frontier/pkg/logger"
)

func MigrateFrontier(logger *slog.Logger, appConfig *config.Frontier) error {
	return cmd.RunMigrations(logger, appConfig.DB)
}

func StartFrontier(logger *slog.Logger, appConfig *config.Frontier) {
	go func() {
		if err := cmd.StartServer(logger, appConfig); err != nil {
			frontierlogger.Fatal(logger, "err starting", "err", err)
			panic(err)
		}
	}()
}
