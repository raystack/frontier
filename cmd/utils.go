package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
	"gopkg.in/yaml.v3"
)

func parseFile(filePath string, v interface{}) error {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	switch filepath.Ext(filePath) {
	case ".json":
		if err := json.Unmarshal(b, v); err != nil {
			return fmt.Errorf("invalid json: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(b, v); err != nil {
			return fmt.Errorf("invalid yaml: %w", err)
		}
	default:
		return errors.New("unsupported file type")
	}

	return nil
}

func setCtxHeader(ctx context.Context, header string) context.Context {
	s := strings.Split(header, ":")
	key := s[0]
	val := s[1]

	md := metadata.New(map[string]string{key: val})
	ctx = metadata.NewOutgoingContext(ctx, md)

	return ctx
}

// func CheckAuth(args []string) bool {
// 	if md, ok := metadata.FromOutgoingContext(ctx); ok {
// 		return len(md) > 0
// 	}
// 	return false
// }

func IsAuthCheckEnabled(cmd *cobra.Command) bool {
	switch cmd.Name() {
	case "help", "config", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd:
		return false
	}

	for c := cmd; c.Parent() != nil; c = c.Parent() {
		if c.Annotations != nil && c.Annotations["skipAuth"] == "true" {
			return false
		}
	}

	return true
}

func IsClientCLI(cmd *cobra.Command) bool {
	for c := cmd; c.Parent() != nil; c = c.Parent() {
		if c.Annotations != nil && c.Annotations["client"] == "true" {
			return true
		}
	}
	return false
}

func IsClientConfigHostExist(cmd *cobra.Command) bool {
	host, err := cmd.Flags().GetString("host")
	fmt.Println(host)
	if err != nil {
		return false
	}
	if host != "" {
		return true
	}
	return false
}

func IsClientConfigHeaderExist(cmd *cobra.Command) bool {
	host, err := cmd.Flags().GetString("header")
	fmt.Println(host)
	if err != nil {
		return false
	}
	if host != "" {
		return true
	}
	return false
}
