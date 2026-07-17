package cmd

import (
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/raystack/frontier/internal/reconcile"
	"github.com/raystack/frontier/proto/v1beta1/frontierv1beta1connect"
	cli "github.com/spf13/cobra"
)

func ReconcileCommand(cliConfig *Config) *cli.Command {
	var (
		filePath string
		dryRun   bool
		header   string
	)
	cmd := &cli.Command{
		Use:   "reconcile",
		Short: "Reconcile platform configuration to a desired-state file",
		Long: heredoc.Doc(`
			Make platform resources match a desired-state YAML file, through the admin API.

			Kinds: PlatformUser (platform admins and members), Permission (custom
			permissions), Role (platform-level roles), Preference (platform
			settings), and Webhook (webhook endpoints). Deleting a permission, a
				custom role, or a webhook needs an explicit
			'delete: true' on its entry; nothing is deleted by omission, and a predefined
			role cannot be deleted. A preference left out of the file resets to its
			default. Log in as a superuser (for example the bootstrap service account)
			with --header.

			Use "frontier export <kind>" to print the current state in this file format.
		`),
		Example: heredoc.Doc(`
			$ frontier reconcile -f platform-users.yaml --dry-run -H "Authorization:Basic <base64>"
			$ frontier reconcile -f platform-users.yaml -H "Authorization:Basic <base64>"
		`),
		Annotations: map[string]string{
			"group":  "core",
			"client": "true",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			data, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("read desired-state file: %w", err)
			}
			registry, err := buildReconcileRegistry(cliConfig.Host, header)
			if err != nil {
				return err
			}
			reports, runErr := reconcile.Run(cmd.Context(), registry, data, dryRun)
			for _, rep := range reports {
				printReconcileReport(cmd, rep)
			}
			return runErr
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the desired-state YAML file")
	mustMarkRequired(cmd, "file")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the plan without applying changes")
	cmd.Flags().StringVarP(&header, "header", "H", "", "Header <key>:<value> for auth, e.g. 'Authorization:Basic <base64>'")
	bindFlagsFromClientConfig(cmd)
	return cmd
}

// reconcileAPI joins the two generated clients: reads live on FrontierService,
// writes on AdminService. The embedded method sets combine to satisfy the
// per-kind API interfaces.
type reconcileAPI struct {
	frontierv1beta1connect.AdminServiceClient
	frontierv1beta1connect.FrontierServiceClient
}

// buildReconcileRegistry holds every reconcilable kind. New kinds register here.
func buildReconcileRegistry(host, header string) (map[string]reconcile.Reconciler, error) {
	adminClient, err := createAdminClient(host)
	if err != nil {
		return nil, err
	}
	frontierClient, err := createClient(host)
	if err != nil {
		return nil, err
	}
	api := reconcileAPI{AdminServiceClient: adminClient, FrontierServiceClient: frontierClient}
	return map[string]reconcile.Reconciler{
		reconcile.KindPlatformUser: reconcile.NewPlatformUserReconciler(adminClient, header),
		reconcile.KindPermission:   reconcile.NewPermissionReconciler(api, header),
		reconcile.KindRole:         reconcile.NewRoleReconciler(api, header),
		reconcile.KindPreference:   reconcile.NewPreferenceReconciler(api, header),
		reconcile.KindWebhook:      reconcile.NewWebhookReconciler(adminClient, header),
	}, nil
}

// printReconcileReport writes the report to stdout. cobra's cmd.Printf falls
// back to stderr, which breaks piping and redirecting the plan.
func printReconcileReport(cmd *cli.Command, rep reconcile.Report) {
	// a failed document yields a zero-value report; there is nothing to print
	if rep.Kind == "" {
		return
	}
	out := cmd.OutOrStdout()
	if len(rep.Planned) == 0 {
		fmt.Fprintf(out, "%s: no changes\n", rep.Kind)
		return
	}
	verb, count := "applied", rep.Applied
	if rep.DryRun {
		verb, count = "planned", len(rep.Planned)
	}
	fmt.Fprintf(out, "%s (%s %d):\n", rep.Kind, verb, count)
	for _, p := range rep.Planned {
		fmt.Fprintf(out, "  - %s\n", p)
	}
}
