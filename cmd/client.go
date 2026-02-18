package cmd

import (
	"fmt"
	"net/http"
	"strings"

	frontierv1beta1connect "github.com/raystack/frontier/proto/v1beta1/frontierv1beta1connect"
	"github.com/spf13/cobra"
)

func createClient(host string) (frontierv1beta1connect.FrontierServiceClient, error) {
	if host == "" {
		return nil, ErrClientConfigHostNotFound
	}
	return frontierv1beta1connect.NewFrontierServiceClient(http.DefaultClient, ensureHTTPScheme(host)), nil
}

func createAdminClient(host string) (frontierv1beta1connect.AdminServiceClient, error) {
	if host == "" {
		return nil, ErrClientConfigHostNotFound
	}
	return frontierv1beta1connect.NewAdminServiceClient(http.DefaultClient, ensureHTTPScheme(host)), nil
}

func ensureHTTPScheme(host string) string {
	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		return host
	}
	return fmt.Sprintf("https://%s", host)
}

func isClientCLI(cmd *cobra.Command) bool {
	for c := cmd; c.Parent() != nil; c = c.Parent() {
		if c.Annotations != nil && c.Annotations["client"] == "true" {
			return true
		}
	}
	return false
}

func overrideClientConfigHost(cmd *cobra.Command, cliConfig *Config) error {
	if cliConfig == nil {
		return ErrClientConfigNotFound
	}

	host, err := cmd.Flags().GetString("host")
	if err == nil && host != "" {
		cliConfig.Host = host
		return nil
	}

	if cliConfig.Host == "" {
		return ErrClientConfigHostNotFound
	}

	return nil
}

func bindFlagsFromClientConfig(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("host", "h", "", "Frontier API service to connect to")
}
