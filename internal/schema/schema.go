package schema

import (
	"context"
	"fmt"
	"strings"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/policy"
	"github.com/odpf/shield/core/role"

	"golang.org/x/exp/maps"
)

type NamespaceType string

var (
	SystemNamespace        NamespaceType = "system_namespace"
	ResourceGroupNamespace NamespaceType = "resource_group_namespace"
)

type NamespaceConfig struct {
	InheritedNamespaces []string
	Type                NamespaceType
	Roles               map[string][]string
	Permissions         map[string][]string
}

type NamespaceConfigMapType map[string]NamespaceConfig

type NamespaceService interface {
	Create(ctx context.Context, ns namespace.Namespace) (namespace.Namespace, error)
}

type RoleService interface {
	Create(ctx context.Context, toCreate role.Role) (role.Role, error)
}

type PolicyService interface {
	Create(ctx context.Context, policy policy.Policy) ([]policy.Policy, error)
}

type ActionService interface {
	Create(ctx context.Context, action action.Action) (action.Action, error)
}

type FileService interface {
	GetSchema(ctx context.Context) (NamespaceConfigMapType, error)
}

type SchemaService struct {
	schemaConfig     FileService
	namespaceService NamespaceService
	roleService      RoleService
	actionService    ActionService
	policyService    PolicyService
}

func NewSchemaMigrationService(
	schemaConfig FileService,
	namespaceService NamespaceService,
	roleService RoleService,
	actionService ActionService,
	policyService PolicyService) *SchemaService {
	return &SchemaService{
		schemaConfig:     schemaConfig,
		namespaceService: namespaceService,
		roleService:      roleService,
		actionService:    actionService,
		policyService:    policyService,
	}
}

//RunMigrations will read NamespaceConfigMap from SchemaConfig blob store and:
//- Create Role, Action, Policy & Namespace in Postgres store
//- Push schema to SpiceDB
func (s SchemaService) RunMigrations(ctx context.Context) error {
	namespaceConfigMap, err := s.schemaConfig.GetSchema(ctx)
	if err != nil {
		return err
	}

	// combining predefined and configured namespaces
	namespaceConfigMap = MergeNamespaceConfigMap(PreDefinedSystemNamespaceConfig, namespaceConfigMap)

	// adding predefined roles and permissions for resource group namespaces
	for n, nc := range namespaceConfigMap {
		if nc.Type == ResourceGroupNamespace {
			namespaceConfigMap = MergeNamespaceConfigMap(namespaceConfigMap, NamespaceConfigMapType{
				n: PreDefinedResourceGroupNamespaceConfig,
			})
		}
	}

	//spiceDBSchema := GenerateSchema(namespaceConfigMap)

	// iterate over namespace
	for namespaceId, v := range namespaceConfigMap {
		// create namespace
		backend := ""
		resourceType := ""
		if v.Type == ResourceGroupNamespace {
			st := strings.Split(namespaceId, "/")
			backend = st[0]
			resourceType = st[1]
		}
		_, err := s.namespaceService.Create(ctx, namespace.Namespace{
			ID:           namespaceId,
			Name:         namespaceId,
			Backend:      backend,
			ResourceType: resourceType,
		})
		if err != nil {
			return err
		}

		// create roles
		for roleId, principals := range v.Roles {
			_, err := s.roleService.Create(ctx, role.Role{
				ID:          fmt.Sprintf("%s:%s", namespaceId, roleId),
				Name:        roleId,
				Types:       principals,
				NamespaceID: namespaceId,
			})
			if err != nil {
				return err
			}
		}

		// create actions
		// IMP: we should depreciate actions with principals
		for actionId := range v.Permissions {
			_, err := s.actionService.Create(ctx, action.Action{
				ID:          fmt.Sprintf("%s.%s", actionId, namespaceId),
				Name:        actionId,
				NamespaceID: namespaceId,
			})
			if err != nil {
				return err
			}
		}
	}

	for namespaceId, v := range namespaceConfigMap {
		// create policies
		for actionId, roles := range v.Permissions {
			for _, r := range roles {
				transformedRole, err := getRoleAndPrincipal(r, namespaceId)
				if err != nil {
					return err
				}

				if _, ok := namespaceConfigMap[transformedRole.NamespaceID].Roles[transformedRole.ID]; !ok {
					return fmt.Errorf("role %s not associated with namespace: %s", transformedRole.ID, transformedRole.NamespaceID)
				}

				_, err = s.policyService.Create(ctx, policy.Policy{
					RoleID:      fmt.Sprintf("%s:%s", transformedRole.NamespaceID, transformedRole.ID),
					NamespaceID: namespaceId,
					ActionID:    fmt.Sprintf("%s.%s", actionId, namespaceId),
				})
				if err != nil {
					return err
				}
			}
		}
	}

	spiceDBSchema := GenerateSchema(namespaceConfigMap)
	fmt.Println(strings.Join(spiceDBSchema, "\n"))

	return nil
}

func MergeNamespaceConfigMap(smallMap, largeMap NamespaceConfigMapType) NamespaceConfigMapType {
	combinedMap := make(NamespaceConfigMapType)
	maps.Copy(combinedMap, smallMap)
	for namespaceName, namespaceConfig := range largeMap {
		if _, ok := combinedMap[namespaceName]; !ok {
			combinedMap[namespaceName] = NamespaceConfig{
				Roles:       make(map[string][]string),
				Permissions: make(map[string][]string),
			}
		}

		for roleName := range namespaceConfig.Roles {
			if _, ok := combinedMap[namespaceName].Roles[roleName]; !ok {
				combinedMap[namespaceName].Roles[roleName] = namespaceConfig.Roles[roleName]
			} else {
				combinedMap[namespaceName].Roles[roleName] = AppendIfUnique(namespaceConfig.Roles[roleName], combinedMap[namespaceName].Roles[roleName])
			}
		}

		for permissionName := range namespaceConfig.Permissions {
			combinedMap[namespaceName].Permissions[permissionName] = AppendIfUnique(namespaceConfig.Permissions[permissionName], combinedMap[namespaceName].Permissions[permissionName])
		}

		if value, ok := combinedMap[namespaceName]; ok {
			value.Type = namespaceConfig.Type
			combinedMap[namespaceName] = value
		}
	}

	return combinedMap
}
