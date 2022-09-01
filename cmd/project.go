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

func ProjectCommand(cliConfig *Config) *cli.Command {
	cmd := &cli.Command{
		Use:     "project",
		Aliases: []string{"projects"},
		Short:   "Manage projects",
		Long: heredoc.Doc(`
			Work with projects.
		`),
		Example: heredoc.Doc(`
			$ shield project create
			$ shield project edit
			$ shield project view
			$ shield project list
		`),
		Annotations: map[string]string{
			"group":  "core",
			"client": "true",
		},
	}

	cmd.AddCommand(createProjectCommand(cliConfig))
	cmd.AddCommand(editProjectCommand(cliConfig))
	cmd.AddCommand(viewProjectCommand(cliConfig))
	cmd.AddCommand(listProjectCommand(cliConfig))

	bindFlagsFromClientConfig(cmd)

	return cmd
}

func createProjectCommand(cliConfig *Config) *cli.Command {
	var filePath, header string

	cmd := &cli.Command{
		Use:   "create",
		Short: "Create a project",
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
			res, err := client.CreateProject(ctx, &shieldv1beta1.CreateProjectRequest{
				Body: &reqBody,
			})
			if err != nil {
				return err
			}

			spinner.Stop()
			fmt.Printf("successfully created project %s with id %s\n", res.GetProject().GetName(), res.GetProject().GetId())
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the project body file")
	cmd.MarkFlagRequired("file")
	cmd.Flags().StringVarP(&header, "header", "H", "", "Header <key>:<value>")
	cmd.MarkFlagRequired("header")

	return cmd
}

func editProjectCommand(cliConfig *Config) *cli.Command {
	var filePath string

	cmd := &cli.Command{
		Use:   "edit",
		Short: "Edit a project",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ shield project edit <project-id> --file=<project-body>
		`),
		Annotations: map[string]string{
			"project:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody shieldv1beta1.ProjectRequestBody
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

			projectID := args[0]
			_, err = client.UpdateProject(cmd.Context(), &shieldv1beta1.UpdateProjectRequest{
				Id:   projectID,
				Body: &reqBody,
			})
			if err != nil {
				return err
			}

			spinner.Stop()
			fmt.Printf("successfully edited project with id %s\n", projectID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the project body file")
	cmd.MarkFlagRequired("file")

	return cmd
}

func viewProjectCommand(cliConfig *Config) *cli.Command {
	var metadata bool

	cmd := &cli.Command{
		Use:   "view",
		Short: "View a project",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ shield project view <project-id>
		`),
		Annotations: map[string]string{
			"project:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client, cancel, err := createClient(cmd.Context(), cliConfig.Host)
			if err != nil {
				return err
			}
			defer cancel()

			projectID := args[0]
			res, err := client.GetProject(cmd.Context(), &shieldv1beta1.GetProjectRequest{
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

func listProjectCommand(cliConfig *Config) *cli.Command {
	cmd := &cli.Command{
		Use:   "list",
		Short: "List all projects",
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

			client, cancel, err := createClient(cmd.Context(), cliConfig.Host)
			if err != nil {
				return err
			}
			defer cancel()

			res, err := client.ListProjects(cmd.Context(), &shieldv1beta1.ListProjectsRequest{})
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
