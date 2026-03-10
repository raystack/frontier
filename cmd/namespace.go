package cmd

import (
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/cli/printer"
	cli "github.com/spf13/cobra"
)

func NamespaceCommand(cliConfig *Config) *cli.Command {
	cmd := &cli.Command{
		Use:     "namespace",
		Aliases: []string{"namespaces"},
		Short:   "Manage namespaces",
		Long: heredoc.Doc(`
			Work with namespaces.
		`),
		Example: heredoc.Doc(`
			$ frontier namespace list
			$ frontier namespace view
		`),
		Annotations: map[string]string{
			"group":  "core",
			"client": "true",
		},
	}

	cmd.AddCommand(listNamespaceCommand(cliConfig))
	cmd.AddCommand(viewNamespaceCommand(cliConfig))

	bindFlagsFromClientConfig(cmd)

	return cmd
}

func viewNamespaceCommand(cliConfig *Config) *cli.Command {
	var header string
	cmd := &cli.Command{
		Use:   "view",
		Short: "View a namespace",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ frontier namespace view <namespace-id>
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

			namespaceID := args[0]
			req, err := newRequest(&frontierv1beta1.GetNamespaceRequest{
				Id: namespaceID,
			}, header)
			if err != nil {
				return err
			}
			res, err := client.GetNamespace(cmd.Context(), req)
			if err != nil {
				return err
			}

			report := [][]string{}

			namespace := res.Msg.GetNamespace()

			spinner.Stop()

			report = append(report, []string{"ID", "NAME", "CREATED AT", "UPDATED AT"})
			report = append(report, []string{
				namespace.GetId(),
				namespace.GetName(),
				namespace.GetCreatedAt().AsTime().String(),
				namespace.GetUpdatedAt().AsTime().String(),
			})
			printer.Table(os.Stdout, report)

			return nil
		},
	}

	cmd.Flags().StringVarP(&header, "header", "H", "", "Header <key>:<value>")

	return cmd
}

func listNamespaceCommand(cliConfig *Config) *cli.Command {
	var header string
	cmd := &cli.Command{
		Use:   "list",
		Short: "List all namespaces",
		Args:  cli.NoArgs,
		Example: heredoc.Doc(`
			$ frontier namespace list
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

			req, err := newRequest(&frontierv1beta1.ListNamespacesRequest{}, header)
			if err != nil {
				return err
			}
			res, err := client.ListNamespaces(cmd.Context(), req)
			if err != nil {
				return err
			}

			report := [][]string{}
			namespaces := res.Msg.GetNamespaces()

			spinner.Stop()

			fmt.Printf(" \nShowing %d namespaces\n \n", len(namespaces))

			report = append(report, []string{"ID", "NAME", "CREATED AT", "UPDATED AT"})
			for _, n := range namespaces {
				report = append(report, []string{
					n.GetId(),
					n.GetName(),
					n.GetCreatedAt().AsTime().String(),
					n.GetUpdatedAt().AsTime().String(),
				})
			}
			printer.Table(os.Stdout, report)

			return nil
		},
	}

	cmd.Flags().StringVarP(&header, "header", "H", "", "Header <key>:<value>")

	return cmd
}
