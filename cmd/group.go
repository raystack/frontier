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

func GroupCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {

	cmd := &cli.Command{
		Use:     "group",
		Aliases: []string{"group"},
		Short:   "Manage group",
		Long: heredoc.Doc(`
			Work with groups.
		`),
		Example: heredoc.Doc(`
			$ shield group create
			$ shield group update
			$ shield group get
			$ shield group list
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
	}

	cmd.AddCommand(createGroupCommand(logger, appConfig))
	cmd.AddCommand(updateGroupCommand(logger, appConfig))
	cmd.AddCommand(getGroupCommand(logger, appConfig))
	cmd.AddCommand(listGroupCommand(logger, appConfig))

	return cmd
}

func createGroupCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	var filePath, header string

	cmd := &cli.Command{
		Use:   "create",
		Short: "create all groups",
		Args:  cli.NoArgs,
		Example: heredoc.Doc(`
			$ shield group create --file=<group-body> --header=<key>:<value>
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody shieldv1beta1.GroupRequestBody
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

			res, err := client.CreateGroup(ctx, &shieldv1beta1.CreateGroupRequest{
				Body: &reqBody,
			})
			if err != nil {
				return err
			}

			spinner.Stop()
			logger.Info(fmt.Sprintf("successfully created group %s with id %s", res.GetGroup().GetName(), res.GetGroup().GetId()))
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the group body file")
	cmd.MarkFlagRequired("file")
	cmd.Flags().StringVarP(&header, "header", "H", "", "Header <key>:<value>")
	cmd.MarkFlagRequired("header")

	return cmd
}

func updateGroupCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	var filePath string

	cmd := &cli.Command{
		Use:   "update",
		Short: "update all groups",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ shield group update <group-id> --file=<group-body>
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody shieldv1beta1.GroupRequestBody
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

			groupID := args[0]
			_, err = client.UpdateGroup(ctx, &shieldv1beta1.UpdateGroupRequest{
				Id:   groupID,
				Body: &reqBody,
			})
			if err != nil {
				return err
			}

			spinner.Stop()
			logger.Info(fmt.Sprintf("successfully updated group with id %s", groupID))
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the group body file")
	cmd.MarkFlagRequired("file")

	return cmd
}

func getGroupCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	var metadata bool

	cmd := &cli.Command{
		Use:   "get",
		Short: "get an group",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ shield group get <group-id>
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

			groupID := args[0]
			res, err := client.GetGroup(ctx, &shieldv1beta1.GetGroupRequest{
				Id: groupID,
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

func listGroupCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {

	cmd := &cli.Command{
		Use:   "list",
		Short: "list all groups",
		Args:  cli.NoArgs,
		Example: heredoc.Doc(`
			$ shield group list
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

			res, err := client.ListGroups(ctx, &shieldv1beta1.ListGroupsRequest{})
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
