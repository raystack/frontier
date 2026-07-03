package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/raystack/frontier/internal/reconcile"
	cli "github.com/spf13/cobra"
)

func ExportCommand(cliConfig *Config) *cli.Command {
	var header string
	cmd := &cli.Command{
		Use:   "export <kind>",
		Short: "Export the current state of a kind as a desired-state file",
		Long: heredoc.Doc(`
			Read the current state of one kind from the server and print it as a
			desired-state YAML document, the format "frontier reconcile" reads. Use it
			to write the first version of an environment's file: reconciling the
			output changes nothing.
		`),
		Example: heredoc.Doc(`
			$ frontier export platformuser -H "Authorization:Basic <base64>"
			$ frontier export platformuser -H "Authorization:Basic <base64>" > platform-users.yaml
		`),
		Annotations: map[string]string{
			"group":  "core",
			"client": "true",
		},
		Args: cli.ExactArgs(1),
		RunE: func(cmd *cli.Command, args []string) error {
			adminClient, err := createAdminClient(cliConfig.Host)
			if err != nil {
				return err
			}
			registry := reconcileRegistry(adminClient, header)
			kind, err := resolveKind(args[0], registry)
			if err != nil {
				return err
			}
			out, err := reconcile.Export(cmd.Context(), registry, kind)
			if err != nil {
				return err
			}
			// cobra's cmd.Print falls back to stderr; the document must go to
			// stdout so redirecting to a file works.
			_, err = fmt.Fprint(cmd.OutOrStdout(), string(out))
			return err
		},
	}
	cmd.Flags().StringVarP(&header, "header", "H", "", "Header <key>:<value> for auth, e.g. 'Authorization:Basic <base64>'")
	bindFlagsFromClientConfig(cmd)
	return cmd
}

// resolveKind matches the kind argument against the registry, case-insensitive
// and accepting a trailing "s", so "platformusers" finds "PlatformUser".
func resolveKind(arg string, registry map[string]reconcile.Reconciler) (string, error) {
	names := make([]string, 0, len(registry))
	for kind := range registry {
		if strings.EqualFold(arg, kind) || strings.EqualFold(arg, kind+"s") {
			return kind, nil
		}
		names = append(names, kind)
	}
	sort.Strings(names)
	return "", fmt.Errorf("unknown kind %q (available: %s)", arg, strings.Join(names, ", "))
}
