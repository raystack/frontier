package bootstrap

import (
	"context"
	"fmt"

	"github.com/raystack/frontier/billing/plan"

	azcore "github.com/authzed/spicedb/pkg/proto/core/v1"

	"github.com/raystack/frontier/core/namespace"
	"github.com/raystack/frontier/core/permission"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/internal/bootstrap/schema"
)

var (
	defaultOrgID = schema.PlatformOrgID.String()
)

type NamespaceService interface {
	Upsert(ctx context.Context, ns namespace.Namespace) (namespace.Namespace, error)
}

type PermissionService interface {
	List(ctx context.Context, flt permission.Filter) ([]permission.Permission, error)
	Upsert(ctx context.Context, action permission.Permission) (permission.Permission, error)
}

type RoleService interface {
	Get(ctx context.Context, id string) (role.Role, error)
	Upsert(ctx context.Context, toCreate role.Role) (role.Role, error)
}

type UserService interface {
	Sudo(ctx context.Context, id string) error
}

type FileService interface {
	GetDefinition(ctx context.Context) (*schema.ServiceDefinition, error)
}

type AuthzEngine interface {
	WriteSchema(ctx context.Context, schema string) error
}

type BillingPlanRepository interface {
	Get(ctx context.Context) (plan.File, error)
}

type PlanService interface {
	UpsertPlans(ctx context.Context, planFile plan.File) error
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

	planService   PlanService
	planLocalRepo BillingPlanRepository
}

func NewBootstrapService(
	config AdminConfig,
	schemaConfig FileService,
	namespaceService NamespaceService,
	roleService RoleService,
	actionService PermissionService,
	userService UserService,
	authzEngine AuthzEngine,
	planService PlanService,
	planLocalRepo BillingPlanRepository,
) *Service {
	return &Service{
		adminConfig:       config,
		schemaConfig:      schemaConfig,
		namespaceService:  namespaceService,
		roleService:       roleService,
		permissionService: actionService,
		userService:       userService,
		authzEngine:       authzEngine,
		planService:       planService,
		planLocalRepo:     planLocalRepo,
	}
}

func (s Service) MigrateSchema(ctx context.Context) error {
	customServiceDefinition, err := s.schemaConfig.GetDefinition(ctx)
	if err != nil {
		return err
	}

	return s.AppendSchema(ctx, *customServiceDefinition)
}

func (s Service) AppendSchema(ctx context.Context, customServiceDefinition schema.ServiceDefinition) error {
	// get existing permissions and append to the new definition
	// this is required to avoid overriding existing permissions in authzed engine
	var existingServiceDefinition schema.ServiceDefinition

	existingPermissions, err := s.permissionService.List(ctx, permission.Filter{})
	if err != nil {
		return nil
	}
	for _, existingPermission := range existingPermissions {
		description := ""
		if existingPermission.Metadata != nil {
			if v, ok := existingPermission.Metadata["description"]; !ok {
				description = v.(string)
			}
		}
		existingServiceDefinition.Permissions = append(existingServiceDefinition.Permissions, schema.ResourcePermission{
			Name:        existingPermission.Name,
			Namespace:   existingPermission.NamespaceID,
			Description: description,
		})
	}

	return s.applySchema(ctx, schema.MergeServiceDefinitions(customServiceDefinition, existingServiceDefinition))
}

// applySchema builds and apply schema over az engine and db
// schema is composed of inbuilt definitions and custom user defined services
// this is idempotent operation and overrides existing schema
func (s Service) applySchema(ctx context.Context, customServiceDefinition *schema.ServiceDefinition) error {
	var err error

	// filter out default app namespace permissions
	customServiceDefinition.Permissions = filterDefaultAppNamespacePermissions(customServiceDefinition.Permissions)

	// build az schema with user defined services
	authzedDefinitions := GetBaseAZSchema()
	authzedDefinitions, err = ApplyServiceDefinitionOverAZSchema(customServiceDefinition, authzedDefinitions)
	if err != nil {
		return fmt.Errorf("MigrateSchema: error applying schema over base: %w", err)
	}

	// validate prepared az schema
	authzedSchemaSource, err := PrepareSchemaAsAZSource(authzedDefinitions)
	if err != nil {
		return fmt.Errorf("PrepareSchemaAsAZSource: %w", err)
	}
	if err = ValidatePreparedAZSchema(ctx, authzedSchemaSource); err != nil {
		return fmt.Errorf("ValidatePreparedAZSchema: %w", err)
	}

	// apply app to db
	appServiceDefinition, err := BuildServiceDefinitionFromAZSchema(authzedDefinitions)
	if err != nil {
		return fmt.Errorf("BuildServiceDefinitionFromAZSchema : %w", err)
	}
	if err = s.migrateAZDefinitionsToDB(ctx, authzedDefinitions); err != nil {
		return fmt.Errorf("migrateAZDefinitionsToDB : %w", err)
	}
	if err = s.migrateServiceDefinitionToDB(ctx, appServiceDefinition); err != nil {
		return fmt.Errorf("migrateServiceDefinitionToDB : %w", err)
	}

	// apply azSchema to engine
	if err = s.authzEngine.WriteSchema(ctx, authzedSchemaSource); err != nil {
		return fmt.Errorf("%w: %s", schema.ErrMigration, err.Error())
	}

	return nil
}

func filterDefaultAppNamespacePermissions(permissions []schema.ResourcePermission) []schema.ResourcePermission {
	var filteredPermissions []schema.ResourcePermission
	for _, permission := range permissions {
		if ns, _ := schema.SplitNamespaceResource(permission.Namespace); ns != schema.DefaultNamespace {
			filteredPermissions = append(filteredPermissions, permission)
		}
	}
	return filteredPermissions
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
	serviceDefinition, err := s.schemaConfig.GetDefinition(ctx)
	if err != nil {
		return err
	}
	for _, defRole := range serviceDefinition.Roles {
		if err = s.createRole(ctx, defaultOrgID, defRole); err != nil {
			return err
		}
	}
	return nil
}

func (s Service) createRole(ctx context.Context, orgID string, defRole schema.RoleDefinition) error {
	if _, err := s.roleService.Get(ctx, defRole.Name); err == nil {
		// role already exists
		return nil
	}

	_, err := s.roleService.Upsert(ctx, role.Role{
		Title:       defRole.Title,
		Name:        defRole.Name,
		OrgID:       orgID,
		Permissions: defRole.Permissions,
		Scopes:      defRole.Scopes,
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
	for _, perm := range appServiceDefinition.Permissions {
		// create permissions if needed
		_, err := s.permissionService.Upsert(ctx, permission.Permission{
			Name:        perm.GetName(),
			NamespaceID: perm.GetNamespace(),
			Metadata: map[string]any{
				"description": perm.Description,
			},
		})
		if err != nil {
			return fmt.Errorf("permissionService.Upsert: %s: %w", err.Error(), schema.ErrMigration)
		}
	}
	return nil
}

// migrateAZDefinitionsToDB will ensure wll the namespaces are already created in database which will be used
// throughout the application
func (s Service) migrateAZDefinitionsToDB(ctx context.Context, azDefinitions []*azcore.NamespaceDefinition) error {
	// iterate over all az definitions and convert frontier namespace
	for _, azDef := range azDefinitions {
		// create namespace if needed
		_, err := s.namespaceService.Upsert(ctx, namespace.Namespace{
			Name: azDef.GetName(),
		})
		if err != nil {
			return fmt.Errorf("namespaceService.Upsert: %w: %s", schema.ErrMigration, err.Error())
		}
	}
	return nil
}

func (s Service) MigrateBillingPlans(ctx context.Context) error {
	localPlans, err := s.planLocalRepo.Get(ctx)
	if err != nil {
		return err
	}

	return s.planService.UpsertPlans(ctx, localPlans)
}
