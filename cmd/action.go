package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/MakeNowJust/heredoc"
	"github.com/odpf/salt/log"
	"github.com/odpf/salt/printer"
	"github.com/odpf/shield/config"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	cli "github.com/spf13/cobra"
)

func ActionCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	cmd := &cli.Command{
		Use:     "action",
		Aliases: []string{"actions"},
		Short:   "Manage actions",
		Long: heredoc.Doc(`
			Work with actions.
		`),
		Example: heredoc.Doc(`
			$ shield action create
			$ shield action edit
			$ shield action view
			$ shield action list
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
	}

	cmd.AddCommand(createActionCommand(logger, appConfig))
	cmd.AddCommand(editActionCommand(logger, appConfig))
	cmd.AddCommand(viewActionCommand(logger, appConfig))
	cmd.AddCommand(listActionCommand(logger, appConfig))

	return cmd
}

func createActionCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	var filePath, header string

	cmd := &cli.Command{
		Use:   "create",
		Short: "Create an action",
		Args:  cli.NoArgs,
		Example: heredoc.Doc(`
			$ shield action create --file=<action-body> --header=<key>:<value>
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody shieldv1beta1.ActionRequestBody
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

			ctx = setCtxHeader(ctx, header)

			res, err := client.CreateAction(ctx, &shieldv1beta1.CreateActionRequest{
				Body: &reqBody,
			})
			if err != nil {
				return err
			}

			spinner.Stop()
			logger.Info(fmt.Sprintf("successfully created action %s with id %s", res.GetAction().GetName(), res.GetAction().GetId()))
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the action body file")
	cmd.MarkFlagRequired("file")
	cmd.Flags().StringVarP(&header, "header", "H", "", "Header <key>:<value>")
	cmd.MarkFlagRequired("header")

	return cmd
}

func editActionCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	var filePath string

	cmd := &cli.Command{
		Use:   "edit",
		Short: "Edit an action",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ shield action edit <action-id> --file=<action-body>
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody shieldv1beta1.ActionRequestBody
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

			actionID := args[0]
			_, err = client.UpdateAction(ctx, &shieldv1beta1.UpdateActionRequest{
				Id:   actionID,
				Body: &reqBody,
			})
			if err != nil {
				return err
			}

			spinner.Stop()
			logger.Info(fmt.Sprintf("successfully edited action with id %s", actionID))
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the action body file")
	cmd.MarkFlagRequired("file")

	return cmd
}

func viewActionCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	cmd := &cli.Command{
		Use:   "view",
		Short: "View an action",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ shield action view <action-id>
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			host := appConfig.App.Host + ":" + strconv.Itoa(appConfig.App.Port)
			ctx := context.Background()
			client, cancel, err := createClient(ctx, host)
			if err != nil {
				return err
			}
			defer cancel()

			actionID := args[0]
			res, err := client.GetAction(ctx, &shieldv1beta1.GetActionRequest{
				Id: actionID,
			})
			if err != nil {
				return err
			}

			report := [][]string{}

			action := res.GetAction()

			spinner.Stop()

			report = append(report, []string{"ID", "NAME", "NAMESPACE"})
			report = append(report, []string{
				action.GetId(),
				action.GetName(),
				action.GetNamespace().GetId(),
			})
			printer.Table(os.Stdout, report)

			return nil
		},
	}

	return cmd
}

func listActionCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	cmd := &cli.Command{
		Use:   "list",
		Short: "List all actions",
		Args:  cli.NoArgs,
		Example: heredoc.Doc(`
			$ shield action list
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			host := appConfig.App.Host + ":" + strconv.Itoa(appConfig.App.Port)
			ctx := context.Background()
			client, cancel, err := createClient(ctx, host)
			if err != nil {
				return err
			}
			defer cancel()

			res, err := client.ListActions(ctx, &shieldv1beta1.ListActionsRequest{})
			if err != nil {
				return err
			}

			report := [][]string{}
			actions := res.GetActions()

			spinner.Stop()

			if len(actions) == 0 {
				fmt.Printf("No actions found.\n")
				return nil
			}

			fmt.Printf(" \nShowing %d action(s)\n \n", len(actions))

			report = append(report, []string{"ID", "NAME", "NAMESPACE"})
			for _, a := range actions {
				report = append(report, []string{
					a.GetId(),
					a.GetName(),
					a.GetNamespace().GetId(),
				})
			}
			printer.Table(os.Stdout, report)

			return nil
		},
	}

	return cmd
}
