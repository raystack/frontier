package main

import (
	"fmt"
	shield_logger "github.com/odpf/shield/logger"
	"os"

	"github.com/odpf/shield/cmd"
	"github.com/odpf/shield/config"
)

func main() {
	appConfig := config.Load()
	logger := shield_logger.InitLogger(appConfig)

	if err := cmd.New(logger, appConfig).Execute(); err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
}
