package testbench

import (
	"bytes"
	"log"

	"github.com/odpf/shield/cmd"
	"github.com/odpf/shield/config"
	shieldlogger "github.com/odpf/shield/pkg/logger"
)

func migrateShield(appConfig *config.Shield) error {
	logger := shieldlogger.InitLogger(appConfig.Log)
	cli := cmd.New(logger, appConfig)

	buf := new(bytes.Buffer)
	cli.SetOutput(buf)
	cli.SetArgs([]string{"migrate"})

	if err := cli.Execute(); err != nil {
		return err
	}

	return nil
}

func startShield(appConfig *config.Shield) {
	logger := shieldlogger.InitLogger(appConfig.Log)
	cli := cmd.New(logger, appConfig)

	buf := new(bytes.Buffer)
	cli.SetOutput(buf)
	cli.SetArgs([]string{"serve"})

	go func() {
		if err := cli.Execute(); err != nil {
			log.Fatal(err)
		}
	}()
}
