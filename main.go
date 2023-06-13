package main

import (
	"fmt"
	"os"

	"github.com/raystack/shield/cmd"

	_ "github.com/authzed/authzed-go/proto/authzed/api/v0"
)

func main() {
	cliConfig, err := cmd.LoadConfig()
	if err != nil {
		cliConfig = &cmd.Config{}
	}
	if err := cmd.New(cliConfig).Execute(); err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
}
