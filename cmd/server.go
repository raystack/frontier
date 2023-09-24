package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/raystack/frontier/pkg/utils"

	"github.com/MakeNowJust/heredoc"
	"github.com/raystack/frontier/config"
	frontierlogger "github.com/raystack/frontier/pkg/logger"
	"github.com/spf13/cobra"
	cli "github.com/spf13/cobra"
)

func ServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "server",
		Aliases: []string{"s"},
		Short:   "Server management",
		Long:    "Server management commands.",
		Example: heredoc.Doc(`
			$ frontier server init
			$ frontier server start
			$ frontier server start -c ./config.yaml
			$ frontier server migrate
			$ frontier server migrate -c ./config.yaml
			$ frontier server migrate-rollback
			$ frontier server migrate-rollback -c ./config.yaml
			$ frontier server keygen
		`),
	}

	cmd.AddCommand(serverInitCommand())
	cmd.AddCommand(serverStartCommand())
	cmd.AddCommand(serverMigrateCommand())
	cmd.AddCommand(serverMigrateRollbackCommand())
	cmd.AddCommand(serverGenRSACommand())

	return cmd
}

func serverInitCommand() *cobra.Command {
	var configFile string
	c := &cli.Command{
		Use:   "init",
		Short: "Initialize server",
		Long: heredoc.Doc(`
			Initializing server. Creating a sample of frontier server config.
			Default: ./config.yaml
		`),
		Example: "frontier server init",
		RunE: func(cmd *cli.Command, args []string) error {
			if err := config.Init(configFile); err != nil {
				return err
			}

			fmt.Printf("server config created: %v\n", configFile)
			return nil
		},
	}

	c.Flags().StringVarP(&configFile, "output", "o", "./config.yaml", "Output config file path")
	return c
}

func serverStartCommand() *cobra.Command {
	var configFile string

	c := &cli.Command{
		Use:     "start",
		Short:   "Start server and proxy default on port 8080",
		Example: "frontier server start",
		RunE: func(cmd *cli.Command, args []string) error {
			appConfig, err := config.Load(configFile)
			if err != nil {
				panic(err)
			}
			logger := frontierlogger.InitLogger(appConfig.Log)

			if err = StartServer(logger, appConfig); err != nil {
				logger.Error("error starting server", "error", err)
				return err
			}
			return nil
		},
	}

	c.Flags().StringVarP(&configFile, "config", "c", "", "config file path")
	return c
}

func serverMigrateCommand() *cobra.Command {
	var configFile string

	c := &cli.Command{
		Use:     "migrate",
		Short:   "Run DB Schema Migrations",
		Example: "frontier server migrate",
		RunE: func(c *cli.Command, args []string) error {
			appConfig, err := config.Load(configFile)
			if err != nil {
				panic(err)
			}

			logger := frontierlogger.InitLogger(appConfig.Log)
			logger.Info("frontier is migrating", "version", config.Version)

			if err = RunMigrations(logger, appConfig.DB); err != nil {
				logger.Error("error running migrations", "error", err)
				return err
			}

			logger.Info("frontier migration complete")
			return nil
		},
	}

	c.Flags().StringVarP(&configFile, "config", "c", "", "config file path")
	return c
}

func serverMigrateRollbackCommand() *cobra.Command {
	var configFile string

	c := &cli.Command{
		Use:     "migrate-rollback",
		Short:   "Run DB Schema Migrations Rollback to last state",
		Example: "frontier migrate-rollback",
		RunE: func(c *cli.Command, args []string) error {
			appConfig, err := config.Load(configFile)
			if err != nil {
				panic(err)
			}
			logger := frontierlogger.InitLogger(appConfig.Log)
			logger.Info("frontier is migrating", "version", config.Version)

			if err = RunRollback(logger, appConfig.DB); err != nil {
				logger.Error("error running migrations rollback", "error", err)
				return err
			}

			logger.Info("frontier migration rollback complete")
			return nil
		},
	}

	c.Flags().StringVarP(&configFile, "config", "c", "", "config file path")
	return c
}

func serverGenRSACommand() *cobra.Command {
	var numOfKeys int
	c := &cli.Command{
		Use:     "keygen",
		Short:   "Generate 2 rsa keys as jwks for auth token generation",
		Example: "frontier server keygen",
		RunE: func(c *cli.Command, args []string) error {
			keySet, err := utils.CreateJWKs(numOfKeys)
			if err != nil {
				return err
			}
			return json.NewEncoder(os.Stdout).Encode(keySet)
		},
	}
	c.Flags().IntVarP(&numOfKeys, "keys", "k", 2, "num of keys to generate")
	return c
}
