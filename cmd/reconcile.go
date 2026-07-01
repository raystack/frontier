package cmd

import (
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/raystack/frontier/internal/reconcile"
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

			Supports the PlatformUser kind (platform admins and members) for now. The file
			decides who has access: anyone listed is added, anyone not listed is removed.
			Log in as a superuser (for example the bootstrap service account) with --header.
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
			adminClient, err := createAdminClient(cliConfig.Host)
			if err != nil {
				return err
			}
			registry := map[string]reconcile.Reconciler{
				reconcile.KindPlatformUser: reconcile.NewPlatformUserReconciler(adminClient, header),
			}
			reports, runErr := reconcile.Run(cmd.Context(), registry, data, dryRun)
			for _, rep := range reports {
				printReconcileReport(cmd, rep)
			}
			return runErr
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the desired-state YAML file")
	cmd.MarkFlagRequired("file")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the plan without applying changes")
	cmd.Flags().StringVarP(&header, "header", "H", "", "Header <key>:<value> for auth, e.g. 'Authorization:Basic <base64>'")
	bindFlagsFromClientConfig(cmd)
	return cmd
}

func printReconcileReport(cmd *cli.Command, rep reconcile.Report) {
	if len(rep.Planned) == 0 {
		cmd.Printf("%s: no changes\n", rep.Kind)
		return
	}
	verb, count := "applied", rep.Applied
	if rep.DryRun {
		verb, count = "planned", len(rep.Planned)
	}
	cmd.Printf("%s (%s %d):\n", rep.Kind, verb, count)
	for _, p := range rep.Planned {
		cmd.Printf("  - %s\n", p)
	}
}
