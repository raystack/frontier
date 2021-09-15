package main

import (
	"fmt"
	"github.com/odpf/shield/cmd"
	"os"
)

func main() {
	if err := cmd.New().Execute(); err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
}
