package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/MakeNowJust/heredoc"
	"github.com/odpf/salt/log"
	"github.com/odpf/salt/printer"
	"github.com/odpf/shield/config"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	cli "github.com/spf13/cobra"
)

func NamespaceCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {

	cmd := &cli.Command{
		Use:     "namespace",
		Aliases: []string{"namespace"},
		Short:   "Manage namespace",
		Long: heredoc.Doc(`
			Work with namespaces.
		`),
		Example: heredoc.Doc(`
			$ shield namespace create
			$ shield namespace update
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
	}

	cmd.AddCommand(createNamespaceCommand(logger, appConfig))
	cmd.AddCommand(updateNamespaceCommand(logger, appConfig))

	return cmd
}

func createNamespaceCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	var filePath string

	cmd := &cli.Command{
		Use:   "create",
		Short: "create all namespaces",
		Args:  cli.NoArgs,
		Example: heredoc.Doc(`
			$ shield namespace create --file=<namespace-body>
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody shieldv1beta1.NamespaceRequestBody
			if err := parseFile(filePath, &reqBody); err != nil {
				return err
			}

			err := reqBody.ValidateAll()
			if err != nil {
				return err
			}

			host := appConfig.App.Host + ":" + strconv.Itoa(appConfig.App.Port)
			ctx := context.Background()
			client, cancel, err := createClient(ctx, host)
			if err != nil {
				return err
			}
			defer cancel()

			res, err := client.CreateNamespace(ctx, &shieldv1beta1.CreateNamespaceRequest{
				Body: &reqBody,
			})
			if err != nil {
				return err
			}

			spinner.Stop()
			logger.Info(fmt.Sprintf("successfully created namespace %s with id %s", res.GetNamespace().GetName(), res.GetNamespace().GetId()))
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the provider config")
	cmd.MarkFlagRequired("file")

	return cmd
}

func updateNamespaceCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	var filePath string

	cmd := &cli.Command{
		Use:   "update",
		Short: "update all namespaces",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ shield namespace update <namespace-id> --file=<namespace-body>
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody shieldv1beta1.NamespaceRequestBody
			if err := parseFile(filePath, &reqBody); err != nil {
				return err
			}

			err := reqBody.ValidateAll()
			if err != nil {
				return err
			}

			host := appConfig.App.Host + ":" + strconv.Itoa(appConfig.App.Port)
			ctx := context.Background()
			client, cancel, err := createClient(ctx, host)
			if err != nil {
				return err
			}
			defer cancel()

			namespaceID := args[0]
			res, err := client.UpdateNamespace(ctx, &shieldv1beta1.UpdateNamespaceRequest{
				Id:   namespaceID,
				Body: &reqBody,
			})
			if err != nil {
				return err
			}

			spinner.Stop()
			logger.Info(fmt.Sprintf("successfully updated namespace with id %s to id %s and name %s", namespaceID, res.GetNamespace().GetId(), res.GetNamespace().GetName()))
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the provider config")
	cmd.MarkFlagRequired("file")

	return cmd
}
