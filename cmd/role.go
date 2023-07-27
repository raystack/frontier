package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/raystack/frontier/pkg/file"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/printer"
	cli "github.com/spf13/cobra"
)

func RoleCommand(cliConfig *Config) *cli.Command {
	cmd := &cli.Command{
		Use:     "role",
		Aliases: []string{"roles"},
		Short:   "Manage roles",
		Long: heredoc.Doc(`
			Work with roles.
		`),
		Example: heredoc.Doc(`
			$ frontier role create
			$ frontier role edit
			$ frontier role view
			$ frontier role list
		`),
		Annotations: map[string]string{
			"group":  "core",
			"client": "true",
		},
	}

	cmd.AddCommand(createRoleCommand(cliConfig))
	cmd.AddCommand(editRoleCommand(cliConfig))
	cmd.AddCommand(viewRoleCommand(cliConfig))
	cmd.AddCommand(listRoleCommand(cliConfig))

	bindFlagsFromClientConfig(cmd)

	return cmd
}

func createRoleCommand(cliConfig *Config) *cli.Command {
	var filePath, header string

	cmd := &cli.Command{
		Use:   "create",
		Short: "Upsert a role",
		Args:  cli.NoArgs,
		Example: heredoc.Doc(`
			$ frontier role create --file=<role-body> --header=<key>:<value>
		`),
		Annotations: map[string]string{
			"role:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody frontierv1beta1.RoleRequestBody
			if err := file.Parse(filePath, &reqBody); err != nil {
				return err
			}

			err := reqBody.ValidateAll()
			if err != nil {
				return err
			}

			client, cancel, err := createClient(cmd.Context(), cliConfig.Host)
			if err != nil {
				return err
			}
			defer cancel()

			ctx := setCtxHeader(cmd.Context(), header)

			res, err := client.CreateOrganizationRole(ctx, &frontierv1beta1.CreateOrganizationRoleRequest{
				Body: &reqBody,
			})
			if err != nil {
				return err
			}

			spinner.Stop()
			fmt.Printf("successfully created role %s with id %s\n", res.GetRole().GetName(), res.GetRole().GetId())
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the role body file")
	cmd.MarkFlagRequired("file")
	cmd.Flags().StringVarP(&header, "header", "H", "", "Header <key>:<value>")
	cmd.MarkFlagRequired("header")

	return cmd
}

func editRoleCommand(cliConfig *Config) *cli.Command {
	var filePath string

	cmd := &cli.Command{
		Use:   "edit",
		Short: "Edit a role",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ frontier role edit <role-id> --file=<role-body>
		`),
		Annotations: map[string]string{
			"role:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody frontierv1beta1.RoleRequestBody
			if err := file.Parse(filePath, &reqBody); err != nil {
				return err
			}

			err := reqBody.ValidateAll()
			if err != nil {
				return err
			}

			client, cancel, err := createClient(cmd.Context(), cliConfig.Host)
			if err != nil {
				return err
			}
			defer cancel()

			roleID := args[0]
			_, err = client.UpdateOrganizationRole(cmd.Context(), &frontierv1beta1.UpdateOrganizationRoleRequest{
				Id:   roleID,
				Body: &reqBody,
			})
			if err != nil {
				return err
			}

			spinner.Stop()
			fmt.Printf("successfully edited role with id %s\n", roleID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the role body file")
	cmd.MarkFlagRequired("file")

	return cmd
}

func viewRoleCommand(cliConfig *Config) *cli.Command {
	var metadata bool

	cmd := &cli.Command{
		Use:   "view",
		Short: "View a role",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ frontier role view <role-id>
		`),
		Annotations: map[string]string{
			"role:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client, cancel, err := createClient(cmd.Context(), cliConfig.Host)
			if err != nil {
				return err
			}
			defer cancel()

			roleID := args[0]
			res, err := client.GetOrganizationRole(cmd.Context(), &frontierv1beta1.GetOrganizationRoleRequest{
				Id: roleID,
			})
			if err != nil {
				return err
			}

			report := [][]string{}

			role := res.GetRole()

			spinner.Stop()

			report = append(report, []string{"ID", "NAME", "PERMISSION(S)", "ORGID"})
			report = append(report, []string{
				role.GetId(),
				role.GetName(),
				strings.Join(role.GetPermissions(), ", "),
				role.GetOrgId(),
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

func listRoleCommand(cliConfig *Config) *cli.Command {
	cmd := &cli.Command{
		Use:   "list",
		Short: "List all roles",
		Args:  cli.NoArgs,
		Example: heredoc.Doc(`
			$ frontier role list
		`),
		Annotations: map[string]string{
			"role:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client, cancel, err := createClient(cmd.Context(), cliConfig.Host)
			if err != nil {
				return err
			}
			defer cancel()

			res, err := client.ListRoles(cmd.Context(), &frontierv1beta1.ListRolesRequest{})
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

			report = append(report, []string{"ID", "NAME", "PERMISSION(S)", "ORGID"})
			for _, r := range roles {
				report = append(report, []string{
					r.GetId(),
					r.GetName(),
					strings.Join(r.GetPermissions(), ", "),
					r.GetOrgId(),
				})
			}
			printer.Table(os.Stdout, report)

			return nil
		},
	}

	return cmd
}
