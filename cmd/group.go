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

func GroupCommand(cliConfig *Config) *cli.Command {
	cmd := &cli.Command{
		Use:     "group",
		Aliases: []string{"groups"},
		Short:   "Manage groups",
		Long: heredoc.Doc(`
			Work with groups.
		`),
		Example: heredoc.Doc(`
			$ shield group create
			$ shield group edit
			$ shield group view
			$ shield group list
		`),
		Annotations: map[string]string{
			"group":  "core",
			"client": "true",
		},
	}

	cmd.AddCommand(createGroupCommand(cliConfig))
	cmd.AddCommand(editGroupCommand(cliConfig))
	cmd.AddCommand(viewGroupCommand(cliConfig))
	cmd.AddCommand(listGroupCommand(cliConfig))

	bindFlagsFromClientConfig(cmd)

	return cmd
}

func createGroupCommand(cliConfig *Config) *cli.Command {
	var filePath, header string

	cmd := &cli.Command{
		Use:   "create",
		Short: "Create a group",
		Args:  cli.NoArgs,
		Example: heredoc.Doc(`
			$ shield group create --file=<group-body> --header=<key>:<value>
		`),
		Annotations: map[string]string{
			"group": "core",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody shieldv1beta1.GroupRequestBody
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

			res, err := client.CreateGroup(setCtxHeader(cmd.Context(), header), &shieldv1beta1.CreateGroupRequest{
				Body: &reqBody,
			})
			if err != nil {
				return err
			}

			spinner.Stop()
			fmt.Printf("successfully created group %s with id %s\n", res.GetGroup().GetName(), res.GetGroup().GetId())
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the group body file")
	cmd.MarkFlagRequired("file")
	cmd.Flags().StringVarP(&header, "header", "H", "", "Header <key>:<value>")
	cmd.MarkFlagRequired("header")

	return cmd
}

func editGroupCommand(cliConfig *Config) *cli.Command {
	var filePath string

	cmd := &cli.Command{
		Use:   "edit",
		Short: "Edit a group",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ shield group edit <group-id> --file=<group-body>
		`),
		Annotations: map[string]string{
			"group": "core",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody shieldv1beta1.GroupRequestBody
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

			groupID := args[0]
			_, err = client.UpdateGroup(cmd.Context(), &shieldv1beta1.UpdateGroupRequest{
				Id:   groupID,
				Body: &reqBody,
			})
			if err != nil {
				return err
			}

			spinner.Stop()
			fmt.Printf("successfully edited group with id %s\n", groupID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the group body file")
	cmd.MarkFlagRequired("file")

	return cmd
}

func viewGroupCommand(cliConfig *Config) *cli.Command {
	var metadata bool

	cmd := &cli.Command{
		Use:   "view",
		Short: "View a group",
		Args:  cli.ExactArgs(2),
		Example: heredoc.Doc(`
			$ shield group view <org-id> <group-id>
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

			orgID := args[0]
			groupID := args[1]
			res, err := client.GetGroup(cmd.Context(), &shieldv1beta1.GetGroupRequest{
				Id:    groupID,
				OrgId: orgID,
			})
			if err != nil {
				return err
			}

			report := [][]string{}

			group := res.GetGroup()

			spinner.Stop()

			report = append(report, []string{"ID", "NAME", "SLUG", "ORG-ID"})
			report = append(report, []string{
				group.GetId(),
				group.GetName(),
				group.GetSlug(),
				group.GetOrgId(),
			})
			printer.Table(os.Stdout, report)

			if metadata {
				meta := group.GetMetadata()
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

func listGroupCommand(cliConfig *Config) *cli.Command {
	cmd := &cli.Command{
		Use:   "list",
		Short: "List all groups",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ shield group list <orgid>
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

			res, err := client.ListOrganizationGroups(cmd.Context(), &shieldv1beta1.ListOrganizationGroupsRequest{
				OrgId: args[0],
			})
			if err != nil {
				return err
			}

			report := [][]string{}
			groups := res.GetGroups()

			spinner.Stop()

			if len(groups) == 0 {
				fmt.Printf("No groups found.\n")
				return nil
			}

			fmt.Printf(" \nShowing %d groups\n \n", len(groups))

			report = append(report, []string{"ID", "NAME", "SLUG", "ORG-ID"})
			for _, g := range groups {
				report = append(report, []string{
					g.GetId(),
					g.GetName(),
					g.GetSlug(),
					g.GetOrgId(),
				})
			}
			printer.Table(os.Stdout, report)

			return nil
		},
	}

	return cmd
}
