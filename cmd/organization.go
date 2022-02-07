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

func OrganizationCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {

	cmd := &cli.Command{
		Use:     "organization",
		Aliases: []string{"organization"},
		Short:   "Manage organization",
		Long: heredoc.Doc(`
			Work with organizations.
		`),
		Example: heredoc.Doc(`
			$ shield organization create
			$ shield organization update
			$ shield organization get
			$ shield organization list
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
	}

	cmd.AddCommand(createOrganizationCommand(logger, appConfig))
	cmd.AddCommand(updateOrganizationCommand(logger, appConfig))
	cmd.AddCommand(getOrganizationCommand(logger, appConfig))
	cmd.AddCommand(listOrganizationCommand(logger, appConfig))
	//cmd.AddCommand(admaddOrganizationCommand(logger, appConfig))
	//cmd.AddCommand(admremoveOrganizationCommand(logger, appConfig))
	//cmd.AddCommand(admlistOrganizationCommand(logger, appConfig))

	return cmd
}

func createOrganizationCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	var filePath, header string

	cmd := &cli.Command{
		Use:   "create",
		Short: "create all organizations",
		Args:  cli.NoArgs,
		Example: heredoc.Doc(`
			$ shield organization create --file=<organization-body> --header=<key>:<value>
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody shieldv1beta1.OrganizationRequestBody
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

			res, err := client.CreateOrganization(ctx, &shieldv1beta1.CreateOrganizationRequest{
				Body: &reqBody,
			})
			if err != nil {
				return err
			}

			spinner.Stop()
			logger.Info(fmt.Sprintf("successfully created organization %s with id %s", res.GetOrganization().GetName(), res.GetOrganization().GetId()))
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the organization body file")
	cmd.MarkFlagRequired("file")
	cmd.Flags().StringVarP(&header, "header", "H", "", "Header <key>:<value>")
	cmd.MarkFlagRequired("header")

	return cmd
}

func updateOrganizationCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	var filePath string

	cmd := &cli.Command{
		Use:   "update",
		Short: "update all organizations",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ shield organization update <organization-id> --file=<organization-body>
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody shieldv1beta1.OrganizationRequestBody
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

			organizationID := args[0]
			_, err = client.UpdateOrganization(ctx, &shieldv1beta1.UpdateOrganizationRequest{
				Id:   organizationID,
				Body: &reqBody,
			})
			if err != nil {
				return err
			}

			spinner.Stop()
			logger.Info(fmt.Sprintf("successfully updated organization with id %s", organizationID))
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the organization body file")
	cmd.MarkFlagRequired("file")

	return cmd
}

func getOrganizationCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	var metadata bool

	cmd := &cli.Command{
		Use:   "get",
		Short: "get an organization",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ shield organization get <organization-id>
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

			organizationID := args[0]
			res, err := client.GetOrganization(ctx, &shieldv1beta1.GetOrganizationRequest{
				Id: organizationID,
			})
			if err != nil {
				return err
			}

			report := [][]string{}

			organization := res.GetOrganization()

			spinner.Stop()

			report = append(report, []string{"ID", "NAME", "SLUG"})
			report = append(report, []string{
				organization.GetId(),
				organization.GetName(),
				organization.GetSlug(),
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

func listOrganizationCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {

	cmd := &cli.Command{
		Use:   "list",
		Short: "list all organizations",
		Args:  cli.NoArgs,
		Example: heredoc.Doc(`
			$ shield organization list
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

			res, err := client.ListOrganizations(ctx, &shieldv1beta1.ListOrganizationsRequest{})
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

			report = append(report, []string{"ID", "NAME", "SLUG"})
			for _, o := range organizations {
				report = append(report, []string{
					o.GetId(),
					o.GetName(),
					o.GetSlug(),
				})
			}
			printer.Table(os.Stdout, report)

			return nil
		},
	}

	return cmd
}

/*
func admaddOrganizationCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	var filePath string

	cmd := &cli.Command{
		Use:   "admadd",
		Short: "add admins to all organizations",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ shield organization admadd <organization-id> -file=<add-organization-admin-body>
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody shieldv1beta1.AddOrganizationAdminRequestBody
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

			organizationID := args[0]
			_, err = client.AddOrganizationAdmin(ctx, &shieldv1beta1.AddOrganizationAdminRequest{
				Id:   organizationID,
				Body: &reqBody,
			})
			if err != nil {
				return err
			}

			spinner.Stop()
			logger.Info(fmt.Sprintf("successfully added admin(s) to organization"))
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the provider config")
	cmd.MarkFlagRequired("file")

	return cmd
}

func admremoveOrganizationCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	var userID string

	cmd := &cli.Command{
		Use:   "admremove",
		Short: "remove admins from all organizations",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ shield organization admremove <organization-id> --user=<user-id>
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

			organizationID := args[0]
			_, err = client.RemoveOrganizationAdmin(ctx, &shieldv1beta1.RemoveOrganizationAdminRequest{
				Id:     organizationID,
				UserId: userID,
			})
			if err != nil {
				return err
			}

			spinner.Stop()
			logger.Info(fmt.Sprintf("successfully removed admin from organization"))
			return nil
		},
	}

	cmd.Flags().StringVarP(&userID, "user", "u", "", "Id of the user to be removed")
	cmd.MarkFlagRequired("user")

	return cmd
}

func admlistOrganizationCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {

	cmd := &cli.Command{
		Use:   "admlist",
		Short: "list admins of all organizations",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ shield organization admlist <organization-id>
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

			organizationID := args[0]
			res, err := client.ListOrganizationAdmins(ctx, &shieldv1beta1.ListOrganizationAdminsRequest{
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
*/
