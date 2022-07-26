package main

import (
	"fmt"
	"os"

	shieldlogger "github.com/odpf/shield/pkg/logger"

	"github.com/odpf/shield/cmd"
	"github.com/odpf/shield/config"

	_ "github.com/authzed/authzed-go/proto/authzed/api/v0"
)

func main() {
	appConfig, err := config.Load()
	if err != nil {
		panic(err)
	}
	logger := shieldlogger.InitLogger(appConfig.Log)

	if err := cmd.New(logger, appConfig).Execute(); err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
}
