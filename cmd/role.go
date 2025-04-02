package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/goto/salt/log"
	"github.com/goto/salt/printer"
	"github.com/odpf/shield/config"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	cli "github.com/spf13/cobra"
)

func RoleCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	cmd := &cli.Command{
		Use:     "role",
		Aliases: []string{"roles"},
		Short:   "Manage roles",
		Long: heredoc.Doc(`
			Work with roles.
		`),
		Example: heredoc.Doc(`
			$ shield role create
			$ shield role edit
			$ shield role view
			$ shield role list
		`),
		Annotations: map[string]string{
			"role:core": "true",
		},
	}

	cmd.AddCommand(createRoleCommand(logger, appConfig))
	cmd.AddCommand(editRoleCommand(logger, appConfig))
	cmd.AddCommand(viewRoleCommand(logger, appConfig))
	cmd.AddCommand(listRoleCommand(logger, appConfig))

	return cmd
}

func createRoleCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	var filePath, header string

	cmd := &cli.Command{
		Use:   "create",
		Short: "Create a role",
		Args:  cli.NoArgs,
		Example: heredoc.Doc(`
			$ shield role create --file=<role-body> --header=<key>:<value>
		`),
		Annotations: map[string]string{
			"role:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody shieldv1beta1.RoleRequestBody
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

			res, err := client.CreateRole(ctx, &shieldv1beta1.CreateRoleRequest{
				Body: &reqBody,
			})
			if err != nil {
				return err
			}

			spinner.Stop()
			logger.Info(fmt.Sprintf("successfully created role %s with id %s", res.GetRole().GetName(), res.GetRole().GetId()))
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the role body file")
	cmd.MarkFlagRequired("file")
	cmd.Flags().StringVarP(&header, "header", "H", "", "Header <key>:<value>")
	cmd.MarkFlagRequired("header")

	return cmd
}

func editRoleCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	var filePath string

	cmd := &cli.Command{
		Use:   "edit",
		Short: "Edit a role",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ shield role edit <role-id> --file=<role-body>
		`),
		Annotations: map[string]string{
			"role:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody shieldv1beta1.RoleRequestBody
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

			roleID := args[0]
			_, err = client.UpdateRole(ctx, &shieldv1beta1.UpdateRoleRequest{
				Id:   roleID,
				Body: &reqBody,
			})
			if err != nil {
				return err
			}

			spinner.Stop()
			logger.Info(fmt.Sprintf("successfully edited role with id %s", roleID))
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the role body file")
	cmd.MarkFlagRequired("file")

	return cmd
}

func viewRoleCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	var metadata bool

	cmd := &cli.Command{
		Use:   "view",
		Short: "View a role",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ shield role view <role-id>
		`),
		Annotations: map[string]string{
			"role:core": "true",
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

			roleID := args[0]
			res, err := client.GetRole(ctx, &shieldv1beta1.GetRoleRequest{
				Id: roleID,
			})
			if err != nil {
				return err
			}

			report := [][]string{}

			role := res.GetRole()

			spinner.Stop()

			report = append(report, []string{"ID", "NAME", "TYPE(S)", "NAMESPACE"})
			report = append(report, []string{
				role.GetId(),
				role.GetName(),
				strings.Join(role.GetTypes(), ", "),
				role.GetNamespace().GetId(),
			})
			printer.Table(os.Stdout, report)

			if metadata {
				meta := role.GetMetadata()
				if len(meta.AsMap()) == 0 {
					fmt.Println("\nNo metadata found")
					return nil
				}

				fmt.Print("\nMETADATA\n")
				metaReport := [][]string{}
				metaReport = append(metaReport, []string{"KEY", "VALUE"})

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

func listRoleCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	cmd := &cli.Command{
		Use:   "list",
		Short: "List all roles",
		Args:  cli.NoArgs,
		Example: heredoc.Doc(`
			$ shield role list
		`),
		Annotations: map[string]string{
			"role:core": "true",
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

			res, err := client.ListRoles(ctx, &shieldv1beta1.ListRolesRequest{})
			if err != nil {
				return err
			}

			report := [][]string{}
			roles := res.GetRoles()

			spinner.Stop()

			if len(roles) == 0 {
				fmt.Printf("No roles found.\n")
				return nil
			}

			fmt.Printf(" \nShowing %d roles\n \n", len(roles))

			report = append(report, []string{"ID", "NAME", "TYPE(S)", "NAMESPACE"})
			for _, r := range roles {
				report = append(report, []string{
					r.GetId(),
					r.GetName(),
					strings.Join(r.GetTypes(), ", "),
					r.GetNamespace().GetId(),
				})
			}
			printer.Table(os.Stdout, report)

			return nil
		},
	}

	return cmd
}
