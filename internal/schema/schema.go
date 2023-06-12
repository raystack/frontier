package schema

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/raystack/shield/core/action"
	"github.com/raystack/shield/core/namespace"
	"github.com/raystack/shield/core/policy"
	"github.com/raystack/shield/core/role"

	"golang.org/x/exp/maps"
)

type NamespaceType string

var (
	SystemNamespace        NamespaceType = "system_namespace"
	ResourceGroupNamespace NamespaceType = "resource_group_namespace"

	ErrMigration = errors.New("error in migrating authz schema")
)

type InheritedNamespace struct {
	Name        string
	NamespaceId string
}

type NamespaceConfig struct {
	InheritedNamespaces []InheritedNamespace
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

type AuthzEngine interface {
	WriteSchema(ctx context.Context, schema NamespaceConfigMapType) error
}

type SchemaService struct {
	schemaConfig     FileService
	namespaceService NamespaceService
	roleService      RoleService
	actionService    ActionService
	policyService    PolicyService
	authzEngine      AuthzEngine
}

func NewSchemaMigrationService(
	schemaConfig FileService,
	namespaceService NamespaceService,
	roleService RoleService,
	actionService ActionService,
	policyService PolicyService,
	authzEngine AuthzEngine) *SchemaService {
	return &SchemaService{
		schemaConfig:     schemaConfig,
		namespaceService: namespaceService,
		roleService:      roleService,
		actionService:    actionService,
		policyService:    policyService,
		authzEngine:      authzEngine,
	}
}

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
			return fmt.Errorf("%w: %s", ErrMigration, err.Error())
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
				return fmt.Errorf("%w: %s", ErrMigration, err.Error())
			}
		}

		// create role for inherited namespaces
		for _, ins := range v.InheritedNamespaces {
			_, err := s.roleService.Create(ctx, role.Role{
				ID:          fmt.Sprintf("%s:%s", namespaceId, ins.Name),
				Name:        ins.Name,
				Types:       []string{ins.NamespaceId},
				NamespaceID: namespaceId,
			})
			if err != nil {
				return fmt.Errorf("%w: %s", ErrMigration, err.Error())
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
				return fmt.Errorf("%w: %s", ErrMigration, err.Error())
			}
		}
	}

	for namespaceId, v := range namespaceConfigMap {
		// create policies
		for actionId, roles := range v.Permissions {
			for _, r := range roles {
				transformedRole, err := getRoleAndPrincipal(r, namespaceId)
				if err != nil {
					return fmt.Errorf("%w: %s", ErrMigration, err.Error())
				}

				if _, ok := namespaceConfigMap[GetNamespace(transformedRole.NamespaceID)].Roles[transformedRole.ID]; !ok {
					return fmt.Errorf("role %s not associated with namespace: %s", transformedRole.ID, transformedRole.NamespaceID)
				}

				_, err = s.policyService.Create(ctx, policy.Policy{
					RoleID:      GetRoleID(GetNamespace(transformedRole.NamespaceID), transformedRole.ID),
					NamespaceID: namespaceId,
					ActionID:    fmt.Sprintf("%s.%s", actionId, namespaceId),
				})
				if err != nil {
					return fmt.Errorf("%w: %s", ErrMigration, err.Error())
				}
			}
		}
	}

	if err = s.authzEngine.WriteSchema(ctx, namespaceConfigMap); err != nil {
		return fmt.Errorf("%w: %s", ErrMigration, err.Error())
	}

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
			value.InheritedNamespaces = AppendIfUnique(value.InheritedNamespaces, namespaceConfig.InheritedNamespaces)
			combinedMap[namespaceName] = value
		}
	}

	return combinedMap
}
