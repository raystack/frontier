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

func ProjectCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {

	cmd := &cli.Command{
		Use:     "project",
		Aliases: []string{"project"},
		Short:   "Manage project",
		Long: heredoc.Doc(`
			Work with projects.
		`),
		Example: heredoc.Doc(`
			$ shield project create
			$ shield project update
			$ shield project get
			$ shield project list
		`),
		Annotations: map[string]string{
			"project:core": "true",
		},
	}

	cmd.AddCommand(createProjectCommand(logger, appConfig))
	cmd.AddCommand(updateProjectCommand(logger, appConfig))
	cmd.AddCommand(getProjectCommand(logger, appConfig))
	cmd.AddCommand(listProjectCommand(logger, appConfig))

	return cmd
}

func createProjectCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	var filePath, header string

	cmd := &cli.Command{
		Use:   "create",
		Short: "create all projects",
		Args:  cli.NoArgs,
		Example: heredoc.Doc(`
			$ shield project create --file=<project-body> --header=<key>:<value>
		`),
		Annotations: map[string]string{
			"project:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody shieldv1beta1.ProjectRequestBody
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

			res, err := client.CreateProject(ctx, &shieldv1beta1.CreateProjectRequest{
				Body: &reqBody,
			})
			if err != nil {
				return err
			}

			spinner.Stop()
			logger.Info(fmt.Sprintf("successfully created project %s with id %s", res.GetProject().GetName(), res.GetProject().GetId()))
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the project body file")
	cmd.MarkFlagRequired("file")
	cmd.Flags().StringVarP(&header, "header", "H", "", "Header <key>:<value>")
	cmd.MarkFlagRequired("header")

	return cmd
}

func updateProjectCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	var filePath string

	cmd := &cli.Command{
		Use:   "update",
		Short: "update all projects",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ shield project update <project-id> --file=<project-body>
		`),
		Annotations: map[string]string{
			"project:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody shieldv1beta1.ProjectRequestBody
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

			projectID := args[0]
			_, err = client.UpdateProject(ctx, &shieldv1beta1.UpdateProjectRequest{
				Id:   projectID,
				Body: &reqBody,
			})
			if err != nil {
				return err
			}

			spinner.Stop()
			logger.Info(fmt.Sprintf("successfully updated project with id %s", projectID))
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the project body file")
	cmd.MarkFlagRequired("file")

	return cmd
}

func getProjectCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	var metadata bool

	cmd := &cli.Command{
		Use:   "get",
		Short: "get an project",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ shield project get <project-id>
		`),
		Annotations: map[string]string{
			"project:core": "true",
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

			projectID := args[0]
			res, err := client.GetProject(ctx, &shieldv1beta1.GetProjectRequest{
				Id: projectID,
			})
			if err != nil {
				return err
			}

			report := [][]string{}

			project := res.GetProject()

			spinner.Stop()

			report = append(report, []string{"ID", "NAME", "SLUG", "ORG-ID"})
			report = append(report, []string{
				project.GetId(),
				project.GetName(),
				project.GetSlug(),
				project.GetOrgId(),
			})
			printer.Table(os.Stdout, report)

			if metadata {
				meta := project.GetMetadata()
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

func listProjectCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {

	cmd := &cli.Command{
		Use:   "list",
		Short: "list all projects",
		Args:  cli.NoArgs,
		Example: heredoc.Doc(`
			$ shield project list
		`),
		Annotations: map[string]string{
			"project:core": "true",
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

			res, err := client.ListProjects(ctx, &shieldv1beta1.ListProjectsRequest{})
			if err != nil {
				return err
			}

			report := [][]string{}
			projects := res.GetProjects()

			spinner.Stop()

			if len(projects) == 0 {
				fmt.Printf("No projects found.\n")
				return nil
			}

			fmt.Printf(" \nShowing %d project(s)\n \n", len(projects))

			report = append(report, []string{"ID", "NAME", "SLUG", "ORG-ID"})
			for _, p := range projects {
				report = append(report, []string{
					p.GetId(),
					p.GetName(),
					p.GetSlug(),
					p.GetOrgId(),
				})
			}
			printer.Table(os.Stdout, report)

			return nil
		},
	}

	return cmd
}
