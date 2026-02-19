package cmd

import (
	"fmt"
	"os"

	"connectrpc.com/connect"
	"github.com/MakeNowJust/heredoc"
	"github.com/raystack/frontier/pkg/file"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/cli/printer"
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
			$ frontier organization create
			$ frontier organization edit
			$ frontier organization view
			$ frontier organization list
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
			$ frontier organization create --file=<organization-body> --header=<key>:<value>
		`),
		Annotations: map[string]string{
			"group": "core",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody frontierv1beta1.OrganizationRequestBody
			if err := file.Parse(filePath, &reqBody); err != nil {
				return err
			}

			err := reqBody.ValidateAll()
			if err != nil {
				return err
			}

			client, err := createClient(cliConfig.Host)
			if err != nil {
				return err
			}

			req, err := newRequest(&frontierv1beta1.CreateOrganizationRequest{
				Body: &reqBody,
			}, header)
			if err != nil {
				return err
			}
			res, err := client.CreateOrganization(cmd.Context(), req)
			if err != nil {
				return err
			}

			spinner.Stop()
			fmt.Printf("successfully created organization %s with id %s\n", res.Msg.GetOrganization().GetName(), res.Msg.GetOrganization().GetId())
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
			$ frontier organization edit <organization-id> --file=<organization-body>
		`),
		Annotations: map[string]string{
			"group": "core",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody frontierv1beta1.OrganizationRequestBody
			if err := file.Parse(filePath, &reqBody); err != nil {
				return err
			}

			err := reqBody.ValidateAll()
			if err != nil {
				return err
			}

			client, err := createClient(cliConfig.Host)
			if err != nil {
				return err
			}

			organizationID := args[0]
			_, err = client.UpdateOrganization(cmd.Context(), connect.NewRequest(&frontierv1beta1.UpdateOrganizationRequest{
				Id:   organizationID,
				Body: &reqBody,
			}))
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
			$ frontier organization view <organization-id>
		`),
		Annotations: map[string]string{
			"group": "core",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client, err := createClient(cliConfig.Host)
			if err != nil {
				return err
			}

			organizationID := args[0]
			res, err := client.GetOrganization(cmd.Context(), connect.NewRequest(&frontierv1beta1.GetOrganizationRequest{
				Id: organizationID,
			}))
			if err != nil {
				return err
			}

			report := [][]string{}

			organization := res.Msg.GetOrganization()

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
			$ frontier organization list
		`),
		Annotations: map[string]string{
			"group": "core",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client, err := createClient(cliConfig.Host)
			if err != nil {
				return err
			}

			res, err := client.ListOrganizations(cmd.Context(), connect.NewRequest(&frontierv1beta1.ListOrganizationsRequest{}))
			if err != nil {
				return err
			}

			report := [][]string{}
			organizations := res.Msg.GetOrganizations()

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
			$ frontier organization admlist <organization-id>
		`),
		Annotations: map[string]string{
			"group": "core",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client, err := createClient(cliConfig.Host)
			if err != nil {
				return err
			}

			organizationID := args[0]
			res, err := client.ListOrganizationAdmins(cmd.Context(), connect.NewRequest(&frontierv1beta1.ListOrganizationAdminsRequest{
				Id: organizationID,
			}))
			if err != nil {
				return err
			}

			report := [][]string{}
			admins := res.Msg.GetUsers()

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
