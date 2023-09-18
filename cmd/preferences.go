package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/printer"
	"github.com/spf13/cobra"
	cli "github.com/spf13/cobra"
)

func PreferencesCommand(cliConfig *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "preferences",
		Aliases: []string{"p"},
		Short:   "Preferences management",
		Long:    "Preferences management commands.",
		Example: heredoc.Doc(`
		$ frontier preferences list
		$ frontier preferences set
		$ frontier preferences describe
		`),
	}

	cmd.AddCommand(preferencesListCommand(cliConfig))
	cmd.AddCommand(preferencesSetCommand(cliConfig))
	cmd.AddCommand(preferencesDescribeCommand(cliConfig))

	bindFlagsFromClientConfig(cmd)

	return cmd
}

func preferencesListCommand(cliConfig *Config) *cobra.Command {
	var header string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list preferences",
		Long:  "List preferences prints the current preferences set in the db. If no preferences are set, it will print the default preferences.",
		Args:  cobra.NoArgs,
		Example: heredoc.Doc(`
			$ frontier preferences list
		`),
		Annotations: map[string]string{
			"group": "core",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody frontierv1beta1.ListPreferencesRequest

			adminClient, cancel, err := createAdminClient(cmd.Context(), cliConfig.Host)
			if err != nil {
				return err
			}
			defer cancel()

			ctx := setCtxHeader(cmd.Context(), header)
			res, err := adminClient.ListPreferences(ctx, &reqBody)
			if err != nil {
				return err
			}

			if len(res.Preferences) == 0 {
				spinner.Stop()
				fmt.Println("No preferences set")
				return nil
			}

			report := [][]string{}
			report = append(report, []string{"Name", "Value", "ResourceType", "ResourceID"})
			for _, preference := range res.Preferences {
				report = append(report, []string{preference.Name, preference.Value, preference.ResourceType, preference.ResourceId})
			}
			spinner.Stop()
			printer.Table(os.Stdout, report)
			return nil
		},
	}

	cmd.Flags().StringVarP(&header, "header", "H", "", "Header <key>:<value>")
	cmd.MarkFlagRequired("header")

	return cmd
}

func preferencesSetCommand(cliConfig *Config) *cobra.Command {
	var header, name, value string
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set preferences",
		Args:  cobra.NoArgs,
		Example: heredoc.Doc(`
			$ frontier preferences set mail_link_subject:Your Frontier login link
			$ frontier preferences set disable_orgs_on_create:true
		`),
		Annotations: map[string]string{
			"group": "core",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			if name == "" || value == "" {
				return fmt.Errorf("name and value are required")
			}

			var reqBody frontierv1beta1.CreatePreferencesRequest
			reqBody.Preferences = append(reqBody.Preferences, &frontierv1beta1.PreferenceRequestBody{
				Name:  name,
				Value: value,
			})

			client, cancel, err := createAdminClient(cmd.Context(), cliConfig.Host)
			if err != nil {
				return err
			}
			defer cancel()

			ctx := setCtxHeader(cmd.Context(), header)
			res, err := client.CreatePreferences(ctx, &reqBody)
			if err != nil {
				return err
			}

			spinner.Stop()
			fmt.Printf("Preference %s set successfully with value: %s \n", res.GetPreference()[0].Name, res.GetPreference()[0].Value)

			return nil
		},
	}

	cmd.Flags().StringVarP(&header, "header", "H", "", "Header <key>:<value>")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of the preference")
	cmd.Flags().StringVarP(&value, "value", "v", "", "Value of the preference")

	cmd.MarkFlagRequired("header")
	return cmd
}

func preferencesDescribeCommand(cliConfig *Config) *cobra.Command {
	var header string
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Traits management",
		Long:  "Display the predefined preferences traits used by Frontier for settings at platform, org, group and project levels.",
		Args:  cobra.NoArgs,
		Example: heredoc.Doc(`
			$ frontier preferences describe
		`),
		Annotations: map[string]string{
			"group": "core",
		},
		RunE: func(cmd *cli.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody frontierv1beta1.DescribePreferencesRequest

			client, cancel, err := createClient(cmd.Context(), cliConfig.Host)
			if err != nil {
				return err
			}
			defer cancel()

			ctx := setCtxHeader(cmd.Context(), header)
			res, err := client.DescribePreferences(ctx, &reqBody)
			if err != nil {
				return err
			}

			spinner.Stop()
			fmt.Println(prettyPrint(res.GetTraits()))

			return nil
		},
	}
	cmd.Flags().StringVarP(&header, "header", "H", "", "Header <key>:<value>")
	cmd.MarkFlagRequired("header")

	return cmd
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}
