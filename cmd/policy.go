package cmd

import (
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/raystack/frontier/pkg/file"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/printer"
	cli "github.com/spf13/cobra"
)

func PolicyCommand(cliConfig *Config) *cli.Command {
	cmd := &cli.Command{
		Use:     "policy",
		Aliases: []string{"policies"},
		Short:   "Manage policies",
		Long: heredoc.Doc(`
			Work with policies.
		`),
		Example: heredoc.Doc(`
			$ frontier policy create
			$ frontier policy edit
			$ frontier policy view
		`),
		Annotations: map[string]string{
			"group":  "core",
			"client": "true",
		},
	}

	cmd.AddCommand(createPolicyCommand(cliConfig))
	cmd.AddCommand(editPolicyCommand(cliConfig))
	cmd.AddCommand(viewPolicyCommand(cliConfig))

	bindFlagsFromClientConfig(cmd)

	return cmd
}

func createPolicyCommand(cliConfig *Config) *cli.Command {
	var filePath, header string

	cmd := &cli.Command{
		Use:   "create",
		Short: "Upsert a policy",
		Args:  cli.NoArgs,
		Example: heredoc.Doc(`
			$ frontier policy create --file=<policy-body> --header=<key>:<value>
		`),
		Annotations: map[string]string{
			"policy:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody frontierv1beta1.PolicyRequestBody
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
			_, err = client.CreatePolicy(ctx, &frontierv1beta1.CreatePolicyRequest{
				Body: &reqBody,
			})
			if err != nil {
				return err
			}

			spinner.Stop()
			fmt.Println("successfully created policy")
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the policy body file")
	cmd.MarkFlagRequired("file")
	cmd.Flags().StringVarP(&header, "header", "H", "", "Header <key>:<value>")
	cmd.MarkFlagRequired("header")

	return cmd
}

func editPolicyCommand(cliConfig *Config) *cli.Command {
	var filePath string

	cmd := &cli.Command{
		Use:   "edit",
		Short: "Edit a policy",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ frontier policy edit <policy-id> --file=<policy-body>
		`),
		Annotations: map[string]string{
			"policy:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody frontierv1beta1.PolicyRequestBody
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

			policyID := args[0]
			_, err = client.UpdatePolicy(cmd.Context(), &frontierv1beta1.UpdatePolicyRequest{
				Id:   policyID,
				Body: &reqBody,
			})
			if err != nil {
				return err
			}

			spinner.Stop()
			fmt.Println("successfully edited policy")
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the policy body file")
	cmd.MarkFlagRequired("file")

	return cmd
}

func viewPolicyCommand(cliConfig *Config) *cli.Command {
	cmd := &cli.Command{
		Use:   "view",
		Short: "View a policy",
		Args:  cli.ExactArgs(1),
		Example: heredoc.Doc(`
			$ frontier policy view <policy-id>
		`),
		Annotations: map[string]string{
			"policy:core": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client, cancel, err := createClient(cmd.Context(), cliConfig.Host)
			if err != nil {
				return err
			}
			defer cancel()

			policyID := args[0]
			res, err := client.GetPolicy(cmd.Context(), &frontierv1beta1.GetPolicyRequest{
				Id: policyID,
			})
			if err != nil {
				return err
			}

			report := [][]string{}

			policy := res.GetPolicy()

			spinner.Stop()

			report = append(report, []string{"ID", "RESOURCE", "PRINCIPAL", "ROLEID"})
			report = append(report, []string{
				policy.GetId(),
				policy.GetResource(),
				policy.GetPrincipal(),
				policy.GetRoleId(),
			})
			printer.Table(os.Stdout, report)

			return nil
		},
	}

	return cmd
}
