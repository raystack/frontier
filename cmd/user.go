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

func UserCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	cmd := &cli.Command{
		Use:     "user",
		Aliases: []string{"users"},
		Short:   "Manage users",
		Long: heredoc.Doc(`
			Work with users.
		`),
		Example: heredoc.Doc(`
			$ shield user create
			$ shield user edit
			$ shield user view
			$ shield user list
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
	}

	cmd.AddCommand(createUserCommand(logger, appConfig))
	cmd.AddCommand(editUserCommand(logger, appConfig))
	cmd.AddCommand(viewUserCommand(logger, appConfig))
	cmd.AddCommand(listUserCommand(logger, appConfig))

	return cmd
}

func createUserCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	var filePath string

	cmd := &cli.Command{
		Use:   "create",
		Short: "Create an user",
		Args:  cli.NoArgs,
		Example: heredoc.Doc(`
			$ shield user create --file=<user-body>
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody shieldv1beta1.UserRequestBody
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

			res, err := client.CreateUser(ctx, &shieldv1beta1.CreateUserRequest{
				Body: &reqBody,
			})
			if err != nil {
				return err
			}

			spinner.Stop()
			logger.Info(fmt.Sprintf("successfully created user %s with id %s", res.GetUser().GetName(), res.GetUser().GetId()))
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the user body file")
	cmd.MarkFlagRequired("file")

	return cmd
}

func editUserCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	var filePath string

	cmd := &cli.Command{
		Use:   "edit",
		Short: "Edit an user",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ shield user edit <user-id> --file=<user-body>
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody shieldv1beta1.UserRequestBody
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

			userID := args[0]
			_, err = client.UpdateUser(ctx, &shieldv1beta1.UpdateUserRequest{
				Id:   userID,
				Body: &reqBody,
			})
			if err != nil {
				return err
			}

			spinner.Stop()
			logger.Info(fmt.Sprintf("successfully edited user with id %s", userID))
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the user body file")
	cmd.MarkFlagRequired("file")

	return cmd
}

func viewUserCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	var metadata bool

	cmd := &cli.Command{
		Use:   "view",
		Short: "View an user",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ shield user view <user-id>
		`),
		Annotations: map[string]string{
			"group:core": "true",
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

			userID := args[0]
			res, err := client.GetUser(ctx, &shieldv1beta1.GetUserRequest{
				Id: userID,
			})
			if err != nil {
				return err
			}

			report := [][]string{}

			user := res.GetUser()

			spinner.Stop()

			report = append(report, []string{"ID", "NAME", "EMAIL"})
			report = append(report, []string{
				user.GetId(),
				user.GetName(),
				user.GetEmail(),
			})
			printer.Table(os.Stdout, report)

			if metadata {
				fmt.Print("\nMETADATA\n")

				metaReport := [][]string{}
				metaReport = append(metaReport, []string{"KEY", "VALUE"})
				meta := user.GetMetadata()
				for k, v := range meta.AsMap() {
					metaReport = append(metaReport, []string{k, fmt.Sprint(v)})
				}
				printer.Table(os.Stdout, metaReport)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&metadata, "metadata", "m", false, "Set this flag to see metadata")

	return cmd
}

func listUserCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	cmd := &cli.Command{
		Use:   "list",
		Short: "List all users",
		Args:  cli.NoArgs,
		Example: heredoc.Doc(`
			$ shield user list
		`),
		Annotations: map[string]string{
			"group:core": "true",
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

			res, err := client.ListUsers(ctx, &shieldv1beta1.ListUsersRequest{})
			if err != nil {
				return err
			}

			report := [][]string{}
			users := res.GetUsers()

			spinner.Stop()

			fmt.Printf(" \nShowing %d users\n \n", len(users))

			report = append(report, []string{"ID", "NAME", "EMAIL"})
			for _, u := range users {
				report = append(report, []string{
					u.GetId(),
					u.GetName(),
					u.GetEmail(),
				})
			}
			printer.Table(os.Stdout, report)

			return nil
		},
	}

	return cmd
}
