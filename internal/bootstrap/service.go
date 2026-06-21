package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/raystack/frontier/billing/plan"

	azcore "github.com/authzed/spicedb/pkg/proto/core/v1"

	"github.com/raystack/frontier/core/namespace"
	"github.com/raystack/frontier/core/permission"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/user"
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
	List(ctx context.Context, f role.Filter) ([]role.Role, error)
	Upsert(ctx context.Context, toCreate role.Role) (role.Role, error)
	Update(ctx context.Context, toUpdate role.Role) (role.Role, error)
}

type RelationService interface {
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	Delete(ctx context.Context, rel relation.Relation) error
	List(ctx context.Context, flt relation.Filter) ([]relation.Relation, error)
}

type UserService interface {
	Sudo(ctx context.Context, id string, relationName string) error
	UnSudo(ctx context.Context, id string, relationName string) error
	GetByID(ctx context.Context, id string) (user.User, error)
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

// PolicyService is policy.Service narrowed to what backfill needs. Goes through
// Create so the SpiceDB rolebinding tuples land alongside the row.
type PolicyService interface {
	Create(ctx context.Context, pol policy.Policy) (policy.Policy, error)
	List(ctx context.Context, flt policy.Filter) ([]policy.Policy, error)
}

// ServiceUserCandidate is a service user missing its owning-org policy row.
type ServiceUserCandidate struct {
	ServiceUserID string
	OrgID         string
}

// ServiceUserBackfiller exposes the set-difference query. Narrow on purpose —
// bootstrap shouldn't be able to mutate service users.
type ServiceUserBackfiller interface {
	ListMissingOrgPolicy(ctx context.Context) ([]ServiceUserCandidate, error)
}

// AdminConfig is platform administration configuration
type AdminConfig struct {
	// Users are a list of email-ids/uuids which needs to be promoted as superusers
	// if email is provided and user doesn't exist, user is created by default
	Users []string `yaml:"users" mapstructure:"users"`

	// Authoritative, when true, makes the configured Users list the single source
	// of truth for human superusers: at boot any human holding the platform admin
	// relation that is NOT in Users is demoted. Service accounts are never touched.
	// WARNING: with an empty/misconfigured list this demotes ALL human superusers.
	Authoritative bool `yaml:"authoritative" mapstructure:"authoritative" default:"false"`
}

type Service struct {
	logger            *slog.Logger
	adminConfig       AdminConfig
	schemaConfig      FileService
	namespaceService  NamespaceService
	roleService       RoleService
	permissionService PermissionService
	authzEngine       AuthzEngine
	userService       UserService
	relationService   RelationService
	policyService     PolicyService
	serviceuserRepo   ServiceUserBackfiller
	patDeniedPerms    map[string]struct{}

	planService   PlanService
	planLocalRepo BillingPlanRepository
}

func NewBootstrapService(
	logger *slog.Logger,
	config AdminConfig,
	schemaConfig FileService,
	namespaceService NamespaceService,
	roleService RoleService,
	actionService PermissionService,
	userService UserService,
	authzEngine AuthzEngine,
	relationService RelationService,
	policyService PolicyService,
	serviceuserRepo ServiceUserBackfiller,
	patDeniedPerms map[string]struct{},
	planService PlanService,
	planLocalRepo BillingPlanRepository,
) *Service {
	return &Service{
		logger:            logger,
		adminConfig:       config,
		schemaConfig:      schemaConfig,
		namespaceService:  namespaceService,
		roleService:       roleService,
		permissionService: actionService,
		userService:       userService,
		authzEngine:       authzEngine,
		planService:       planService,
		planLocalRepo:     planLocalRepo,
		relationService:   relationService,
		policyService:     policyService,
		serviceuserRepo:   serviceuserRepo,
		patDeniedPerms:    patDeniedPerms,
	}
}

func (s Service) MigrateSchema(ctx context.Context) error {
	customServiceDefinition, err := s.schemaConfig.GetDefinition(ctx)
	if err != nil {
		return err
	}

	return s.AppendSchema(ctx, *customServiceDefinition)
}

// BuiltinPermissions returns the permissions that come from the base schema and
// the config files — the ones bootstrap recreates on every boot. It looks only
// at the base schema and config, not at the permissions already in the database.
func (s Service) BuiltinPermissions(ctx context.Context) (map[string]struct{}, error) {
	custom, err := s.schemaConfig.GetDefinition(ctx)
	if err != nil {
		return nil, err
	}
	custom.Permissions = filterDefaultAppNamespacePermissions(custom.Permissions)

	defs, err := ApplyServiceDefinitionOverAZSchema(custom, GetBaseAZSchema())
	if err != nil {
		return nil, err
	}
	appDef, err := BuildServiceDefinitionFromAZSchema(defs)
	if err != nil {
		return nil, err
	}

	slugs := make(map[string]struct{}, len(appDef.Permissions))
	for _, p := range appDef.Permissions {
		slugs[schema.FQPermissionNameFromNamespace(p.GetNamespace(), p.GetName())] = struct{}{}
	}
	return slugs, nil
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

// MakeSuperUsers promotes the configured users to superuser. When the admin
// config is Authoritative, it then demotes any human superuser no longer present
// in the config (see reconcileSuperUsers).
func (s Service) MakeSuperUsers(ctx context.Context) error {
	for _, userID := range s.adminConfig.Users {
		userID = strings.TrimSpace(userID)
		slog.DebugContext(ctx, "promoting user to superuser", "user_id", userID)
		if err := s.userService.Sudo(ctx, userID, schema.AdminRelationName); err != nil {
			return err
		}
	}

	if !s.adminConfig.Authoritative {
		return nil
	}
	return s.reconcileSuperUsers(ctx)
}

// reconcileSuperUsers demotes any human user holding the platform admin relation
// that is not present in the configured admin list. Service accounts and the
// member relation are never touched. This makes config the source of truth for
// the human-superuser set; with an empty/misconfigured list it demotes everyone.
func (s Service) reconcileSuperUsers(ctx context.Context) error {
	// Desired set: resolve each config entry to its canonical user ID so the diff
	// is keyed by ID (an admin listed by both email and UUID collapses to one).
	desiredIDs := make(map[string]struct{}, len(s.adminConfig.Users))
	for _, entry := range s.adminConfig.Users {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		u, err := s.userService.GetByID(ctx, entry)
		if err != nil {
			// An entry that doesn't resolve to a user (e.g. a UUID/slug that was
			// never created) cannot be an admin; skip it, matching the promote path.
			slog.WarnContext(ctx, "skipping unresolvable admin config entry during reconciliation", "entry", entry, "err", err.Error())
			continue
		}
		desiredIDs[u.ID] = struct{}{}
	}

	// Current set: every relation on the platform object.
	relations, err := s.relationService.List(ctx, relation.Filter{
		Object: relation.Object{
			ID:        schema.PlatformID,
			Namespace: schema.PlatformNamespace,
		},
	})
	if err != nil {
		return fmt.Errorf("reconciling superusers: listing platform relations: %w", err)
	}

	for _, rel := range relations {
		// Only reconcile human-user admins; leave service accounts and members alone.
		if rel.RelationName != schema.AdminRelationName || rel.Subject.Namespace != schema.UserPrincipal {
			continue
		}
		if _, ok := desiredIDs[rel.Subject.ID]; ok {
			continue
		}
		slog.InfoContext(ctx, "demoting superuser not present in admin config", "user_id", rel.Subject.ID)
		if err := s.userService.UnSudo(ctx, rel.Subject.ID, schema.AdminRelationName); err != nil {
			return fmt.Errorf("reconciling superusers: demoting user %s: %w", rel.Subject.ID, err)
		}
	}
	return nil
}

// MigrateRoles migrate predefined roles to org
func (s Service) MigrateRoles(ctx context.Context) error {
	var err error
	// migrate predefined roles to org
	for _, defRole := range schema.PredefinedRoles {
		if err = s.migrateRole(ctx, defaultOrgID, defRole); err != nil {
			return err
		}
	}

	// migrate user defined roles to org
	serviceDefinition, err := s.schemaConfig.GetDefinition(ctx)
	if err != nil {
		return err
	}
	for _, defRole := range serviceDefinition.Roles {
		if err = s.migrateRole(ctx, defaultOrgID, defRole); err != nil {
			return err
		}
	}

	// backfill PAT wildcard tuples for all existing roles
	if err = s.migratePATRelations(ctx); err != nil {
		return fmt.Errorf("migrating PAT role relations: %w", err)
	}
	return nil
}

// migrateRole makes the stored role match its definition: create when missing,
// reconcile when present. Roles are keyed by name, so renaming a definition
// produces a new role rather than renaming the existing one.
func (s Service) migrateRole(ctx context.Context, orgID string, defRole schema.RoleDefinition) error {
	existing, err := s.roleService.Get(ctx, defRole.Name)
	if errors.Is(err, role.ErrNotExist) {
		return s.createRole(ctx, orgID, defRole)
	}
	if err != nil {
		// A transient Get failure must not fall through to create: that would
		// re-Upsert an existing role and skip the prune, the very drift this fixes.
		return fmt.Errorf("get role %s: %w: %s", defRole.Name, schema.ErrMigration, err.Error())
	}
	return s.reconcileRole(ctx, existing, defRole)
}

func (s Service) createRole(ctx context.Context, orgID string, defRole schema.RoleDefinition) error {
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

// reconcileRole writes the definition onto an existing role only when its
// permission set differs. The write goes through role.Update (not a raw row
// update) because that is what prunes the SpiceDB permission tuples a narrowed
// definition no longer grants; the equality short-circuit keeps every boot from
// rewriting unchanged roles and emitting a spurious RoleUpdated audit record.
func (s Service) reconcileRole(ctx context.Context, existing role.Role, defRole schema.RoleDefinition) error {
	if permissionsEqual(existing.Permissions, defRole.Permissions) {
		return nil
	}
	existing.Title = defRole.Title
	existing.Permissions = defRole.Permissions
	existing.Scopes = defRole.Scopes
	if existing.Metadata == nil {
		existing.Metadata = map[string]any{}
	}
	existing.Metadata["description"] = defRole.Description
	if _, err := s.roleService.Update(ctx, existing); err != nil {
		return fmt.Errorf("can't reconcile role: %w: %s", schema.ErrMigration, err.Error())
	}
	return nil
}

// permissionsEqual reports whether two permission slug sets are equal, ignoring
// order and duplicates.
func permissionsEqual(a, b []string) bool {
	setA := make(map[string]struct{}, len(a))
	for _, p := range a {
		setA[p] = struct{}{}
	}
	setB := make(map[string]struct{}, len(b))
	for _, p := range b {
		setB[p] = struct{}{}
	}
	if len(setA) != len(setB) {
		return false
	}
	for p := range setB {
		if _, ok := setA[p]; !ok {
			return false
		}
	}
	return true
}

// MigrateServiceUserOrgPolicies backfills the org policy for service users that
// have only a SpiceDB member relation (legacy creation flow). Idempotent: on a
// clean cluster the candidate query returns zero rows and this is a no-op.
// Per-row failures are joined into the return value and also logged; the call
// site decides whether to abort or warn-and-continue.
func (s Service) MigrateServiceUserOrgPolicies(ctx context.Context) error {
	candidates, err := s.serviceuserRepo.ListMissingOrgPolicy(ctx)
	if err != nil {
		return fmt.Errorf("list candidates: %w", err)
	}
	if len(candidates) == 0 {
		return nil
	}

	viewerRole, err := s.roleService.Get(ctx, schema.RoleOrganizationViewer)
	if err != nil {
		return fmt.Errorf("get viewer role: %w", err)
	}

	var errs error
	for _, c := range candidates {
		_, createErr := s.policyService.Create(ctx, policy.Policy{
			RoleID:        viewerRole.ID,
			ResourceID:    c.OrgID,
			ResourceType:  schema.OrganizationNamespace,
			PrincipalID:   c.ServiceUserID,
			PrincipalType: schema.ServiceUserPrincipal,
		})
		if createErr != nil {
			errs = errors.Join(errs, fmt.Errorf("backfill SU %s on org %s: %w", c.ServiceUserID, c.OrgID, createErr))
			s.logger.WarnContext(ctx, "backfill failed for service user, continuing",
				"serviceuser_id", c.ServiceUserID,
				"org_id", c.OrgID,
				"error", createErr,
			)
			continue
		}
		s.logger.InfoContext(ctx, "backfilled SU org policy",
			"serviceuser_id", c.ServiceUserID,
			"org_id", c.OrgID,
		)
	}
	return errs
}

// migratePATRelations ensures app/pat:* wildcard tuples are in sync with the current
// denied_permissions config for all existing roles. Runs on every bootstrap:
//   - Creates app/pat:* for allowed permissions (idempotent via SpiceDB Touch)
//   - Deletes app/pat:* for denied permissions (removes stale wildcards if config changed)
func (s Service) migratePATRelations(ctx context.Context) error {
	roles, err := s.roleService.List(ctx, role.Filter{})
	if err != nil {
		return fmt.Errorf("listing roles for PAT migration: %w", err)
	}
	for _, r := range roles {
		for _, perm := range r.Permissions {
			rel := relation.Relation{
				Object: relation.Object{
					ID:        r.ID,
					Namespace: schema.RoleNamespace,
				},
				Subject: relation.Subject{
					ID:        "*",
					Namespace: schema.PATPrincipal,
				},
				RelationName: perm,
			}
			if _, denied := s.patDeniedPerms[perm]; denied {
				if err := s.relationService.Delete(ctx, rel); err != nil {
					return fmt.Errorf("deleting PAT wildcard for role %s denied permission %s: %w", r.Name, perm, err)
				}
				continue
			}
			if _, err := s.relationService.Create(ctx, rel); err != nil {
				return fmt.Errorf("creating PAT wildcard for role %s permission %s: %w", r.Name, perm, err)
			}
		}
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
