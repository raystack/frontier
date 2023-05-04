package schema

import (
	_ "embed"
	"errors"
)

// SpiceDB readable format is stored in predefined_schema.txt
const (
	// Global IDs
	PlatformID = "platform"

	// namespace
	PlatformNamespace     = "app/platform"
	OrganizationNamespace = "app/organization"
	ProjectNamespace      = "app/project"
	GroupNamespace        = "app/group"
	RoleBindingNamespace  = "app/rolebinding"
	RoleNamespace         = "app/role"

	// relation
	PlatformRelationName     = "platform"
	AdminRelationName        = "admin"
	OrganizationRelationName = "org"
	ProjectRelationName      = "project"
	GroupRelationName        = "group"
	MemberRelationName       = "member"
	RoleRelationName         = "role"
	RoleGrantRelationName    = "granted"
	RoleBearerRelationName   = "bearer"

	// relations
	OwnerRelation = "owner"

	// Roles
	OwnerRole  = "owner"
	MemberRole = "member"

	// permissions
	ListPermission          = "list"
	GetPermission           = "get"
	CreatePermission        = "create"
	UpdatePermission        = "update"
	DeletePermission        = "delete"
	SudoPermission          = "superuser"
	RoleManagePermission    = "rolemanage"
	PolicyManagePermission  = "policymanage"
	ProjectListPermission   = "projectlist"
	GroupListPermission     = "grouplist"
	ProjectCreatePermission = "projectcreate"
	GroupCreatePermission   = "groupcreate"
	ResourceListPermission  = "resourcelist"

	// synthetic permission
	MembershipPermission = "membership"

	// principals
	UserPrincipal        = "app/user"
	ServiceUserPrincipal = "app/serviceuser"
	GroupPrincipal       = "app/group"
	SuperUserPrincipal   = "app/superuser"
)

//go:embed base_schema.zed
var BaseSchemaZed string

var (
	ErrMigration = errors.New("error in migrating authz schema")
)

// ServiceDefinition are provided by user for a service
type ServiceDefinition struct {
	Name      string
	Resources []DefinitionResource
}

// DefinitionResource is an object over which authz rules will be applied
type DefinitionResource struct {
	Name        string
	Permissions []ResourcePermission
}

// ResourcePermission with which roles will be created. Whenever an action is performed
// subject access permissions are checked with subject required permissions
type ResourcePermission struct {
	Name        string
	Description string
}

// DefinitionRoles are a set of permissions which can be assigned to a user or group
type DefinitionRoles struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Permissions []string `yaml:"permissions"`
}

type RoleFile struct {
	Roles []DefinitionRoles `yaml:"roles"`
}

var PredefinedRoles = []DefinitionRoles{
	// org
	{
		Name: "app_organization_owner",
		Permissions: []string{
			"app_organization_administer",
		},
	},
	{
		Name: "app_organization_manager",
		Permissions: []string{
			"app_organization_update",
			"app_organization_get",
		},
	},
	{
		Name: "app_organization_viewer",
		Permissions: []string{
			"app_organization_get",
		},
	},
	// project
	{
		Name: "app_project_owner",
		Permissions: []string{
			"app_project_administer",
		},
	},
	{
		Name: "app_project_manager",
		Permissions: []string{
			"app_project_update",
			"app_project_get",
			"app_organization_projectcreate",
			"app_organization_projectlist",
		},
	},
	{
		Name: "app_project_viewer",
		Permissions: []string{
			"app_project_get",
		},
	},
}
