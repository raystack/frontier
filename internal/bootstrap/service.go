package bootstrap

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/permission"
	"github.com/odpf/shield/core/role"
	"github.com/odpf/shield/internal/bootstrap/schema"
)

var (
	defaultOrgID = uuid.Nil.String()
)

type NamespaceService interface {
	Upsert(ctx context.Context, ns namespace.Namespace) (namespace.Namespace, error)
}

type PermissionService interface {
	Upsert(ctx context.Context, action permission.Permission) (permission.Permission, error)
}

type RoleService interface {
	Upsert(ctx context.Context, toCreate role.Role) (role.Role, error)
}

type UserService interface {
	Sudo(ctx context.Context, id string) error
}

type FileService interface {
	GetSchemas(ctx context.Context) ([]schema.ServiceDefinition, error)
	GetRoles(ctx context.Context) ([]schema.DefinitionRoles, error)
}

type AuthzEngine interface {
	WriteSchema(ctx context.Context, schema string) error
}

// AdminConfig is platform administration configuration
type AdminConfig struct {
	// Users are a list of email-ids/uuids which needs to be promoted as superusers
	// if email is provided and user doesn't exist, user is created by default
	Users []string `yaml:"users" mapstructure:"users"`
}

type Service struct {
	adminConfig       AdminConfig
	schemaConfig      FileService
	namespaceService  NamespaceService
	roleService       RoleService
	permissionService PermissionService
	authzEngine       AuthzEngine
	userService       UserService
}

func NewBootstrapService(
	config AdminConfig,
	schemaConfig FileService,
	namespaceService NamespaceService,
	roleService RoleService,
	actionService PermissionService,
	userService UserService,
	authzEngine AuthzEngine) *Service {
	return &Service{
		adminConfig:       config,
		schemaConfig:      schemaConfig,
		namespaceService:  namespaceService,
		roleService:       roleService,
		permissionService: actionService,
		userService:       userService,
		authzEngine:       authzEngine,
	}
}

func (s Service) MigrateSchema(ctx context.Context) error {
	customServiceDefinitions, err := s.schemaConfig.GetSchemas(ctx)
	if err != nil {
		return err
	}

	// build az schema with user defined services
	authzedDefinitions := GetBaseAZSchema()
	for _, serviceDef := range customServiceDefinitions {
		authzedDefinitions, err = ApplyServiceDefinitionOverAZSchema(serviceDef, authzedDefinitions)
		if err != nil {
			return fmt.Errorf("MigrateSchema: error applying schema over base: %w", err)
		}

		if err := s.migrateServiceDefinitionToDB(ctx, serviceDef); err != nil {
			return err
		}
	}

	// validate prepared az schema
	authzedSchemaSource, err := PrepareSchemaAsSource(authzedDefinitions)
	if err != nil {
		return err
	}
	if err = ValidatePreparedAZSchema(ctx, authzedSchemaSource); err != nil {
		return err
	}

	// apply base app to db
	appServiceDefinition, err := BuildServiceDefinitionFromAZSchema("app", authzedDefinitions)
	if err != nil {
		return err
	}
	if err := s.migrateServiceDefinitionToDB(ctx, appServiceDefinition); err != nil {
		return err
	}

	// apply custom service defs to db
	for _, serviceDef := range customServiceDefinitions {
		if err := s.migrateServiceDefinitionToDB(ctx, serviceDef); err != nil {
			return err
		}
	}

	// apply azSchema to engine
	if err = s.authzEngine.WriteSchema(ctx, authzedSchemaSource); err != nil {
		return fmt.Errorf("%w: %s", schema.ErrMigration, err.Error())
	}

	// promote normal users to superusers
	if err = s.MakeSuperUsers(ctx); err != nil {
		return err
	}

	return nil
}

// MakeSuperUsers promote ordinary users to superuser
func (s Service) MakeSuperUsers(ctx context.Context) error {
	for _, userID := range s.adminConfig.Users {
		if err := s.userService.Sudo(ctx, userID); err != nil {
			return err
		}
	}
	return nil
}

// MigrateRoles migrate predefined roles to org
func (s Service) MigrateRoles(ctx context.Context) error {
	var err error
	// migrate predefined roles to org
	for _, defRole := range schema.PredefinedRoles {
		if err = s.createRole(ctx, defaultOrgID, defRole); err != nil {
			return err
		}
	}
	// migrate user defined roles to org
	definitionRoles, err := s.schemaConfig.GetRoles(ctx)
	if err != nil {
		return err
	}
	for _, defRole := range definitionRoles {
		if err = s.createRole(ctx, defaultOrgID, defRole); err != nil {
			return err
		}
	}
	return nil
}

func (s Service) createRole(ctx context.Context, orgID string, defRole schema.DefinitionRoles) error {
	_, err := s.roleService.Upsert(ctx, role.Role{
		Name:        defRole.Name,
		OrgID:       orgID,
		Permissions: defRole.Permissions,
		Metadata: map[string]any{
			"description": defRole.Description,
		},
		State: role.Enabled,
	})
	if err != nil {
		return fmt.Errorf("can't migrate role: %w: %s", schema.ErrMigration, err.Error())
	}
	return nil
}

func (s Service) migrateServiceDefinitionToDB(ctx context.Context, appServiceDefinition schema.ServiceDefinition) error {
	// iterate over definition resources
	for _, resource := range appServiceDefinition.Resources {
		namespaceName := GetNamespaceName(appServiceDefinition.Name, resource.Name)

		// create namespace
		ns, err := s.namespaceService.Upsert(ctx, namespace.Namespace{
			Name: namespaceName,
		})
		if err != nil {
			return fmt.Errorf("namespaceService.Upsert: %w: %s", schema.ErrMigration, err.Error())
		}

		// create permissions
		for _, perm := range resource.Permissions {
			_, err := s.permissionService.Upsert(ctx, permission.Permission{
				Name:        perm.Name,
				NamespaceID: ns.Name,
				Metadata: map[string]any{
					"description": perm.Description,
				},
			})
			if err != nil {
				return fmt.Errorf("permissionService.Upsert: %w: %s", schema.ErrMigration, err.Error())
			}
		}
	}
	return nil
}
