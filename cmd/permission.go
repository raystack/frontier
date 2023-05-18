package cmd

import (
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/odpf/salt/printer"
	"github.com/odpf/shield/pkg/file"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	cli "github.com/spf13/cobra"
)

func PermissionCommand(cliConfig *Config) *cli.Command {
	cmd := &cli.Command{
		Use:     "permission",
		Aliases: []string{"permissions"},
		Short:   "Manage permissions",
		Long: heredoc.Doc(`
			Work with permissions.
		`),
		Example: heredoc.Doc(`
			$ shield permission create
			$ shield permission edit
			$ shield permission view
			$ shield permission list
		`),
		Annotations: map[string]string{
			"group":  "core",
			"client": "true",
		},
	}

	cmd.AddCommand(createPermissionCommand(cliConfig))
	cmd.AddCommand(editPermissionCommand(cliConfig))
	cmd.AddCommand(viewPermissionCommand(cliConfig))
	cmd.AddCommand(listPermissionCommand(cliConfig))

	bindFlagsFromClientConfig(cmd)

	return cmd
}

func createPermissionCommand(cliConfig *Config) *cli.Command {
	var filePath, header string

	cmd := &cli.Command{
		Use:   "create",
		Short: "Upsert a permission",
		Args:  cli.NoArgs,
		Example: heredoc.Doc(`
			$ shield permission create --file=<action-body> --header=<key>:<value>
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody shieldv1beta1.PermissionRequestBody
			if err := file.Parse(filePath, &reqBody); err != nil {
				return err
			}

			err := reqBody.ValidateAll()
			if err != nil {
				return err
			}

			client, cancel, err := createAdminClient(cmd.Context(), cliConfig.Host)
			if err != nil {
				return err
			}
			defer cancel()

			ctx := setCtxHeader(cmd.Context(), header)
			res, err := client.CreatePermission(ctx, &shieldv1beta1.CreatePermissionRequest{
				Bodies: []*shieldv1beta1.PermissionRequestBody{&reqBody},
			})
			if err != nil {
				return err
			}

			spinner.Stop()
			fmt.Printf("successfully created permissions %d\n", len(res.GetPermissions()))
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the permission body file")
	cmd.MarkFlagRequired("file")
	cmd.Flags().StringVarP(&header, "header", "H", "", "Header <key>:<value>")
	cmd.MarkFlagRequired("header")

	return cmd
}

func editPermissionCommand(cliConfig *Config) *cli.Command {
	var filePath string

	cmd := &cli.Command{
		Use:   "edit",
		Short: "Edit a permission",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ shield permission edit <action-id> --file=<action-body>
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody shieldv1beta1.PermissionRequestBody
			if err := file.Parse(filePath, &reqBody); err != nil {
				return err
			}

			err := reqBody.ValidateAll()
			if err != nil {
				return err
			}

			client, cancel, err := createAdminClient(cmd.Context(), cliConfig.Host)
			if err != nil {
				return err
			}
			defer cancel()

			permissionID := args[0]
			_, err = client.UpdatePermission(cmd.Context(), &shieldv1beta1.UpdatePermissionRequest{
				Id:   permissionID,
				Body: &reqBody,
			})
			if err != nil {
				return err
			}

			spinner.Stop()
			fmt.Printf("successfully edited permission with id %s\n", permissionID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the permission body file")
	cmd.MarkFlagRequired("file")

	return cmd
}

func viewPermissionCommand(cliConfig *Config) *cli.Command {
	cmd := &cli.Command{
		Use:   "view",
		Short: "View a permission",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ shield permission view <action-id>
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client, cancel, err := createClient(cmd.Context(), cliConfig.Host)
			if err != nil {
				return err
			}
			defer cancel()

			permissionID := args[0]
			res, err := client.GetPermission(cmd.Context(), &shieldv1beta1.GetPermissionRequest{
				Id: permissionID,
			})
			if err != nil {
				return err
			}

			report := [][]string{}

			action := res.GetPermission()

			spinner.Stop()

			report = append(report, []string{"ID", "NAME", "NAMESPACE"})
			report = append(report, []string{
				action.GetId(),
				action.GetName(),
				action.GetNamespace(),
			})
			printer.Table(os.Stdout, report)

			return nil
		},
	}

	return cmd
}

func listPermissionCommand(cliConfig *Config) *cli.Command {
	cmd := &cli.Command{
		Use:   "list",
		Short: "List all permissions",
		Args:  cli.NoArgs,
		Example: heredoc.Doc(`
			$ shield permission list
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client, cancel, err := createClient(cmd.Context(), cliConfig.Host)
			if err != nil {
				return err
			}
			defer cancel()

			res, err := client.ListPermissions(cmd.Context(), &shieldv1beta1.ListPermissionsRequest{})
			if err != nil {
				return err
			}

			report := [][]string{}
			permissions := res.GetPermissions()

			spinner.Stop()

			if len(permissions) == 0 {
				fmt.Printf("No permissions found.\n")
				return nil
			}

			fmt.Printf(" \nShowing %d permission(s)\n \n", len(permissions))

			report = append(report, []string{"ID", "NAME", "NAMESPACE"})
			for _, a := range permissions {
				report = append(report, []string{
					a.GetId(),
					a.GetName(),
					a.GetNamespace(),
				})
			}
			printer.Table(os.Stdout, report)

			return nil
		},
	}

	return cmd
}
