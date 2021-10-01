package main

import (
	"fmt"
	"os"

	"github.com/odpf/salt/log"
	"github.com/odpf/shield/cmd"
	"github.com/odpf/shield/config"
)

func main() {
	appConfig := config.Load()
	logger := log.NewLogrus(log.LogrusWithLevel(appConfig.Log.Level))

	if err := cmd.New(logger, appConfig).Execute(); err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
}
