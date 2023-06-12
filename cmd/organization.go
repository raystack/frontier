package cmd

import (
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/raystack/salt/printer"
	"github.com/raystack/shield/pkg/file"
	shieldv1beta1 "github.com/raystack/shield/proto/v1beta1"
	cli "github.com/spf13/cobra"
)

func OrganizationCommand(cliConfig *Config) *cli.Command {
	cmd := &cli.Command{
		Use:     "organization",
		Aliases: []string{"organizations"},
		Short:   "Manage organizations",
		Long: heredoc.Doc(`
			Work with organizations.
		`),
		Example: heredoc.Doc(`
			$ shield organization create
			$ shield organization edit
			$ shield organization view
			$ shield organization list
		`),
		Annotations: map[string]string{
			"group":  "core",
			"client": "true",
		},
	}

	cmd.AddCommand(createOrganizationCommand(cliConfig))
	cmd.AddCommand(editOrganizationCommand(cliConfig))
	cmd.AddCommand(viewOrganizationCommand(cliConfig))
	cmd.AddCommand(listOrganizationCommand(cliConfig))
	cmd.AddCommand(admlistOrganizationCommand(cliConfig))

	bindFlagsFromClientConfig(cmd)

	return cmd
}

func createOrganizationCommand(cliConfig *Config) *cli.Command {
	var filePath, header string

	cmd := &cli.Command{
		Use:   "create",
		Short: "Upsert an organization",
		Args:  cli.NoArgs,
		Example: heredoc.Doc(`
			$ shield organization create --file=<organization-body> --header=<key>:<value>
		`),
		Annotations: map[string]string{
			"group": "core",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody shieldv1beta1.OrganizationRequestBody
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
			res, err := client.CreateOrganization(ctx, &shieldv1beta1.CreateOrganizationRequest{
				Body: &reqBody,
			})
			if err != nil {
				return err
			}

			spinner.Stop()
			fmt.Printf("successfully created organization %s with id %s\n", res.GetOrganization().GetName(), res.GetOrganization().GetId())
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the organization body file")
	cmd.MarkFlagRequired("file")
	cmd.Flags().StringVarP(&header, "header", "H", "", "Header <key>:<value>")
	cmd.MarkFlagRequired("header")

	return cmd
}

func editOrganizationCommand(cliConfig *Config) *cli.Command {
	var filePath string

	cmd := &cli.Command{
		Use:   "edit",
		Short: "Edit an organization",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ shield organization edit <organization-id> --file=<organization-body>
		`),
		Annotations: map[string]string{
			"group": "core",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody shieldv1beta1.OrganizationRequestBody
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

			organizationID := args[0]
			_, err = client.UpdateOrganization(cmd.Context(), &shieldv1beta1.UpdateOrganizationRequest{
				Id:   organizationID,
				Body: &reqBody,
			})
			if err != nil {
				return err
			}

			spinner.Stop()
			fmt.Printf("successfully edited organization with id %s\n", organizationID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the organization body file")
	cmd.MarkFlagRequired("file")

	return cmd
}

func viewOrganizationCommand(cliConfig *Config) *cli.Command {
	var metadata bool

	cmd := &cli.Command{
		Use:   "view",
		Short: "View an organization",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ shield organization view <organization-id>
		`),
		Annotations: map[string]string{
			"group": "core",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client, cancel, err := createClient(cmd.Context(), cliConfig.Host)
			if err != nil {
				return err
			}
			defer cancel()

			organizationID := args[0]
			res, err := client.GetOrganization(cmd.Context(), &shieldv1beta1.GetOrganizationRequest{
				Id: organizationID,
			})
			if err != nil {
				return err
			}

			report := [][]string{}

			organization := res.GetOrganization()

			spinner.Stop()

			report = append(report, []string{"ID", "NAME"})
			report = append(report, []string{
				organization.GetId(),
				organization.GetName(),
			})
			printer.Table(os.Stdout, report)

			if metadata {
				meta := organization.GetMetadata()
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

func listOrganizationCommand(cliConfig *Config) *cli.Command {
	cmd := &cli.Command{
		Use:   "list",
		Short: "List all organizations",
		Args:  cli.NoArgs,
		Example: heredoc.Doc(`
			$ shield organization list
		`),
		Annotations: map[string]string{
			"group": "core",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client, cancel, err := createClient(cmd.Context(), cliConfig.Host)
			if err != nil {
				return err
			}
			defer cancel()

			res, err := client.ListOrganizations(cmd.Context(), &shieldv1beta1.ListOrganizationsRequest{})
			if err != nil {
				return err
			}

			report := [][]string{}
			organizations := res.GetOrganizations()

			spinner.Stop()

			if len(organizations) == 0 {
				fmt.Printf("No organizations found.\n")
				return nil
			}

			fmt.Printf(" \nShowing %d organizations\n \n", len(organizations))

			report = append(report, []string{"ID", "NAME"})
			for _, o := range organizations {
				report = append(report, []string{
					o.GetId(),
					o.GetName(),
				})
			}
			printer.Table(os.Stdout, report)

			return nil
		},
	}

	return cmd
}

func admlistOrganizationCommand(cliConfig *Config) *cli.Command {
	cmd := &cli.Command{
		Use:   "admlist",
		Short: "list admins of an organization",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ shield organization admlist <organization-id>
		`),
		Annotations: map[string]string{
			"group": "core",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client, cancel, err := createClient(cmd.Context(), cliConfig.Host)
			if err != nil {
				return err
			}
			defer cancel()

			organizationID := args[0]
			res, err := client.ListOrganizationAdmins(cmd.Context(), &shieldv1beta1.ListOrganizationAdminsRequest{
				Id: organizationID,
			})
			if err != nil {
				return err
			}

			report := [][]string{}
			admins := res.GetUsers()

			spinner.Stop()

			fmt.Printf(" \nShowing %d admins\n \n", len(admins))

			report = append(report, []string{"ID", "NAME", "EMAIL"})
			for _, a := range admins {
				report = append(report, []string{
					a.GetId(),
					a.GetName(),
					a.GetEmail(),
				})
			}
			printer.Table(os.Stdout, report)

			return nil
		},
	}

	return cmd
}
