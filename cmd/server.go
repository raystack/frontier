package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/raystack/shield/pkg/utils"

	"github.com/MakeNowJust/heredoc"
	"github.com/raystack/shield/config"
	shieldlogger "github.com/raystack/shield/pkg/logger"
	"github.com/spf13/cobra"
	cli "github.com/spf13/cobra"
)

// Version of the current build. overridden by the build system.
// see "Makefile" for more information
var (
	Version string
)

func ServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "server",
		Aliases: []string{"s"},
		Short:   "Server management",
		Long:    "Server management commands.",
		Example: heredoc.Doc(`
			$ shield server init
			$ shield server start
			$ shield server start -c ./config.yaml
			$ shield server migrate
			$ shield server migrate -c ./config.yaml
			$ shield server migrate-rollback
			$ shield server migrate-rollback -c ./config.yaml
			$ shield server keygen
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
	var resourcesURL string
	var rulesURL string

	c := &cli.Command{
		Use:   "init",
		Short: "Initialize server",
		Long: heredoc.Doc(`
			Initializing server. Creating a sample of shield server config.
			Default: ./config.yaml
		`),
		Example: "shield server init",
		RunE: func(cmd *cli.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			defaultResourcesURL := fmt.Sprintf("file://%s", path.Join(cwd, "resources_config"))
			defaultRulesURL := fmt.Sprintf("file://%s", path.Join(cwd, "rules"))

			if resourcesURL == "" {
				resourcesURL = defaultResourcesURL
			}
			if rulesURL == "" {
				rulesURL = defaultRulesURL
			}

			if err := config.Init(resourcesURL, rulesURL, configFile); err != nil {
				return err
			}

			fmt.Printf("server config created: %v\n", configFile)
			return nil
		},
	}

	c.Flags().StringVarP(&configFile, "output", "o", "./config.yaml", "Output config file path")
	c.Flags().StringVarP(&resourcesURL, "resources", "r", "", heredoc.Doc(`
		URL path of resources. Full path prefixed with scheme where resources config yaml files are kept
		e.g.:
		local storage file "file:///tmp/resources_config"
		GCS Bucket "gs://shield-bucket-example"
		(default: file://{pwd}/resources_config)
	`))
	c.Flags().StringVarP(&rulesURL, "rule", "u", "", heredoc.Doc(`
		URL path of rules. Full path prefixed with scheme where ruleset yaml files are kept
		e.g.:
		local storage file "file:///tmp/rules"
		GCS Bucket "gs://shield-bucket-example"
		(default: file://{pwd}/rules)
	`))

	return c
}

func serverStartCommand() *cobra.Command {
	var configFile string

	c := &cli.Command{
		Use:     "start",
		Short:   "Start server and proxy default on port 8080",
		Example: "shield server start",
		RunE: func(cmd *cli.Command, args []string) error {
			appConfig, err := config.Load(configFile)
			if err != nil {
				panic(err)
			}
			logger := shieldlogger.InitLogger(appConfig.Log)

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
		Example: "shield server migrate",
		RunE: func(c *cli.Command, args []string) error {
			appConfig, err := config.Load(configFile)
			if err != nil {
				panic(err)
			}

			logger := shieldlogger.InitLogger(appConfig.Log)
			logger.Info("shield is migrating", "version", Version)

			if err = RunMigrations(logger, appConfig.DB); err != nil {
				logger.Error("error running migrations", "error", err)
				return err
			}

			logger.Info("shield migration completed")
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
		Example: "shield migrate-rollback",
		RunE: func(c *cli.Command, args []string) error {
			appConfig, err := config.Load(configFile)
			if err != nil {
				panic(err)
			}
			logger := shieldlogger.InitLogger(appConfig.Log)
			logger.Info("shield is migrating", "version", Version)

			if err = RunRollback(appConfig.DB); err != nil {
				logger.Error("error running migrations rollback", "error", err)
				return err
			}

			logger.Info("shield migration rollback completed")
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
		Example: "shield server keygen",
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
