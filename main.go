package main

import (
	"fmt"
	"os"

	shieldlogger "github.com/odpf/shield/logger"

	"github.com/odpf/shield/cmd"
	"github.com/odpf/shield/config"

	_ "github.com/authzed/authzed-go/proto/authzed/api/v0"
)

func main() {
	appConfig := config.Load()
	logger := shieldlogger.InitLogger(appConfig)

	if err := cmd.New(logger, appConfig).Execute(); err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
}
