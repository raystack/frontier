package blob

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/odpf/shield/internal/schema"

	"github.com/pkg/errors"
	"gocloud.dev/blob"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
)

type RoleConfig struct {
	Name       string   `yaml:"name" json:"name"`
	Principals []string `yaml:"principals" json:"principals"`
}

type PermissionsConfig struct {
	Name  string   `yaml:"name" json:"name"`
	Roles []string `yaml:"roles" json:"roles"`
}

type ResourceTypeConfig struct {
	Name        string              `yaml:"name" json:"name"`
	Roles       []RoleConfig        `yaml:"roles" json:"roles"`
	Permissions []PermissionsConfig `yaml:"permissions" json:"permissions"`
}

type Config struct {
	Type string `yaml:"type" json:"type"`

	ResourceTypes []ResourceTypeConfig `yaml:"resource_types" json:"resource_types,omitempty"`

	Roles       []RoleConfig        `yaml:"roles" json:"roles,omitempty"`
	Permissions []PermissionsConfig `yaml:"permissions" json:"permissions,omitempty"`
}

type SchemaConfig struct {
	bucket Bucket
	config schema.NamespaceConfigMapType
}

func NewSchemaConfigRepository() {

}

func (s *SchemaConfig) GetSchema(ctx context.Context) (schema.NamespaceConfigMapType, error) {
	configMap := make(schema.NamespaceConfigMapType)
	if s.config != nil {
		return s.config, nil
	}

	configFromFiles, err := s.readYAMLFiles(ctx)
	if err != nil {
		return nil, err
	}

	for _, c := range configFromFiles {
		for k, v := range c {
			if v.Type == "resource_group" {
				configMap = mergeNamespaceConfigMap(configMap, getNamespacesForResourceGroup(k, v))
			} else {
				configMap = mergeNamespaceConfigMap(getNamespaceFromConfig(k, v.Roles, v.Permissions), configMap)
			}
		}
	}

	s.config = configMap

	return configMap, nil
}

func mergeNamespaceConfigMap(smallMap, largeMap schema.NamespaceConfigMapType) schema.NamespaceConfigMapType {
	combinedMap := make(schema.NamespaceConfigMapType)
	maps.Copy(combinedMap, smallMap)
	for namespaceName, namespaceConfig := range largeMap {
		if _, ok := combinedMap[namespaceName]; !ok {
			combinedMap[namespaceName] = schema.NamespaceConfig{
				Roles:       make(map[string][]string),
				Permissions: make(map[string][]string),
			}
		}

		for roleName := range namespaceConfig.Roles {
			if _, ok := combinedMap[namespaceName].Roles[roleName]; !ok {
				combinedMap[namespaceName].Roles[roleName] = namespaceConfig.Roles[roleName]
			} else {
				combinedMap[namespaceName].Roles[roleName] = append(namespaceConfig.Roles[roleName], combinedMap[namespaceName].Roles[roleName]...)
			}
		}

		for permissionName := range namespaceConfig.Permissions {
			combinedMap[namespaceName].Permissions[permissionName] = append(namespaceConfig.Permissions[permissionName], combinedMap[namespaceName].Permissions[permissionName]...)
		}
	}

	return combinedMap
}

func (s *SchemaConfig) readYAMLFiles(ctx context.Context) ([]map[string]Config, error) {
	configYAMLs := make([]map[string]Config, 0)

	// iterate over bucket files, only read .yml & .yaml files
	it := s.bucket.List(&blob.ListOptions{})
	for {
		obj, err := it.Next(ctx)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if obj.IsDir {
			continue
		}
		if !(strings.HasSuffix(obj.Key, ".yaml") || strings.HasSuffix(obj.Key, ".yml")) {
			continue
		}
		fileBytes, err := s.bucket.ReadAll(ctx, obj.Key)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", "error in reading bucket object", err.Error())
		}

		var configYAML map[string]Config
		if err := yaml.Unmarshal(fileBytes, &configYAML); err != nil {
			return nil, errors.Wrap(err, "yaml.Unmarshal: "+obj.Key)
		}
		if len(configYAML) == 0 {
			continue
		}

		configYAMLs = append(configYAMLs, configYAML)
	}

	return configYAMLs, nil
}

func getNamespacesForResourceGroup(name string, c Config) schema.NamespaceConfigMapType {
	namespaceConfig := schema.NamespaceConfigMapType{}

	for _, v := range c.ResourceTypes {
		maps.Copy(namespaceConfig, getNamespaceFromConfig(fmt.Sprintf("%s/%s", name, v.Name), v.Roles, v.Permissions))
	}

	return namespaceConfig
}

func getNamespaceFromConfig(name string, rolesConfigs []RoleConfig, permissionConfigs []PermissionsConfig) schema.NamespaceConfigMapType {
	tnc := schema.NamespaceConfig{
		Roles:       make(map[string][]string),
		Permissions: make(map[string][]string),
	}

	for _, v1 := range rolesConfigs {
		tnc.Roles[v1.Name] = v1.Principals
	}

	for _, v2 := range permissionConfigs {
		tnc.Permissions[v2.Name] = v2.Roles
	}

	return schema.NamespaceConfigMapType{name: tnc}
}
