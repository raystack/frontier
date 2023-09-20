package schema

import (
	_ "embed"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// SpiceDB readable format is stored in predefined_schema.txt
const (
	// DefaultNamespace is the default namespace for predefined entities
	DefaultNamespace = "app"

	// Global IDs
	PlatformID = "platform"

	// namespace
	PlatformNamespace     = "app/platform"
	OrganizationNamespace = "app/organization"
	ProjectNamespace      = "app/project"
	GroupNamespace        = "app/group"
	RoleBindingNamespace  = "app/rolebinding"
	RoleNamespace         = "app/role"
	InvitationNamespace   = "app/invitation"

	// relations
	PlatformRelationName     = "platform"
	AdminRelationName        = "admin"
	OrganizationRelationName = "org"
	UserRelationName         = "user"
	ProjectRelationName      = "project"
	GroupRelationName        = "group"
	MemberRelationName       = "member"
	OwnerRelationName        = "owner"
	RoleRelationName         = "role"
	RoleGrantRelationName    = "granted"
	RoleBearerRelationName   = "bearer"

	// permissions
	ListPermission              = "list"
	GetPermission               = "get"
	CreatePermission            = "create"
	UpdatePermission            = "update"
	DeletePermission            = "delete"
	SudoPermission              = "superuser"
	RoleManagePermission        = "rolemanage"
	PolicyManagePermission      = "policymanage"
	ProjectListPermission       = "projectlist"
	GroupListPermission         = "grouplist"
	ProjectCreatePermission     = "projectcreate"
	GroupCreatePermission       = "groupcreate"
	ResourceListPermission      = "resourcelist"
	InvitationListPermission    = "invitationlist"
	InvitationCreatePermission  = "invitationcreate"
	AcceptPermission            = "accept"
	ServiceUserManagePermission = "serviceusermanage"
	ManagePermission            = "manage"

	// synthetic permission
	MembershipPermission = "membership"

	// principals
	UserPrincipal        = "app/user"
	ServiceUserPrincipal = "app/serviceuser"
	GroupPrincipal       = "app/group"
	SuperUserPrincipal   = "app/superuser"

	// Roles
	RoleProjectOwner = "app_project_owner"
	GroupOwnerRole   = "app_group_owner"
	GroupMemberRole  = "app_group_member"
)

var (
	PlatformOrgID = uuid.Nil
)

//go:embed base_schema.zed
var BaseSchemaZed string

var (
	ErrMigration    = errors.New("error in migrating authz schema")
	ErrBadNamespace = errors.New("bad namespace, format should namespace:uuid")
)

// ServiceDefinition is provided by user for a service
type ServiceDefinition struct {
	Roles       []RoleDefinition     `yaml:"roles"`
	Permissions []ResourcePermission `yaml:"permissions"`
}

// MergeServiceDefinitions merges multiple service definitions into one
// and deduplicate roles and permissions
func MergeServiceDefinitions(definitions ...ServiceDefinition) *ServiceDefinition {
	roles := make(map[string]RoleDefinition)
	permissions := make(map[string]ResourcePermission)
	for _, definition := range definitions {
		for _, role := range definition.Roles {
			roles[role.Name] = role
		}
		for _, permission := range definition.Permissions {
			permissions[permission.Slug()] = permission
		}
	}
	roleList := make([]RoleDefinition, 0, len(roles))
	for _, role := range roles {
		roleList = append(roleList, role)
	}
	permissionList := make([]ResourcePermission, 0, len(permissions))
	for _, permission := range permissions {
		permissionList = append(permissionList, permission)
	}
	return &ServiceDefinition{
		Roles:       roleList,
		Permissions: permissionList,
	}
}

// RoleDefinition are a set of permissions which can be assigned to a user or group
type RoleDefinition struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Permissions []string `yaml:"permissions"`
}

// ResourcePermission with which roles will be created. Whenever an action is performed
// subject access permissions are checked with subject required permissions
type ResourcePermission struct {
	// simple name
	Name string

	// Namespace is an object over which authz rules will be applied
	Namespace   string
	Description string

	// Key is a unique identifier composed of namespace and name
	// for example: "app.platform.list" which is composed as service.resource.verb
	// here app.platform is namespace and list is name of the permission
	Key string
}

func (r ResourcePermission) GetName() string {
	if r.Name != "" {
		return r.Name
	}
	_, name := PermissionNamespaceAndNameFromKey(r.Key)
	return name
}

func (r ResourcePermission) GetNamespace() string {
	if r.Namespace != "" {
		return r.Namespace
	}
	ns, _ := PermissionNamespaceAndNameFromKey(r.Key)
	return ns
}

func (r ResourcePermission) Slug() string {
	if r.Key != "" {
		return FQPermissionNameFromNamespace(PermissionNamespaceAndNameFromKey(r.Key))
	}
	return FQPermissionNameFromNamespace(r.Namespace, r.Name)
}

func BuildNamespaceName(service, resource string) string {
	return fmt.Sprintf("%s/%s", service, resource)
}

func SplitNamespaceResource(ns string) (string, string) {
	ns = ParseNamespaceAliasIfRequired(ns)
	parts := strings.Split(ns, "/")
	if len(parts) < 2 {
		return parts[0], "default"
	}
	return parts[0], parts[1]
}

// SplitNamespaceAndResourceID splits ns/something:uuid into ns/something and uuid
func SplitNamespaceAndResourceID(namespace string) (string, string, error) {
	namespaceParts := strings.Split(namespace, ":")
	if len(namespaceParts) != 2 {
		return "", "", ErrBadNamespace
	}

	namespaceName := ParseNamespaceAliasIfRequired(namespaceParts[0])
	resourceID := namespaceParts[1]
	return namespaceName, resourceID, nil
}

func JoinNamespaceAndResourceID(namespace, id string) string {
	return fmt.Sprintf("%s:%s", namespace, id)
}

func ParseNamespaceAliasIfRequired(n string) string {
	switch n {
	case "user":
		n = UserPrincipal
	case "superuser":
		n = SuperUserPrincipal
	case "serviceuser":
		n = ServiceUserPrincipal
	case "group":
		n = GroupPrincipal
	case "org", "organization":
		n = OrganizationNamespace
	case "project":
		n = ProjectNamespace
	}
	return n
}

func FQPermissionNameFromNamespace(namespace, verb string) string {
	service, resource := SplitNamespaceResource(namespace)
	return fmt.Sprintf("%s_%s_%s", service, resource, verb)
}

func PermissionNamespaceAndNameFromKey(key string) (string, string) {
	parts := strings.Split(key, ".")
	if len(parts) != 3 {
		return "", ""
	}
	return fmt.Sprintf("%s/%s", parts[0], parts[1]), parts[2]
}

func PermissionKeyFromNamespaceAndName(namespace, name string) string {
	service, resource := SplitNamespaceResource(namespace)
	return fmt.Sprintf("%s.%s.%s", service, resource, name)
}

func IsSystemNamespace(namespace string) bool {
	return namespace == OrganizationNamespace || namespace == ProjectNamespace ||
		namespace == UserPrincipal || namespace == ServiceUserPrincipal ||
		namespace == SuperUserPrincipal || namespace == GroupPrincipal ||
		namespace == PlatformNamespace
}

// IsValidPermissionName checks if the provided name is a valid permission name
func IsValidPermissionName(name string) bool {
	if name == "" {
		return false
	}
	// check if name contains anything other than alphanumeric characters
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}

var PredefinedRoles = []RoleDefinition{
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
			"app_organization_projectcreate",
			"app_organization_projectlist",
			"app_organization_groupcreate",
			"app_organization_grouplist",
			"app_organization_serviceusermanage",
		},
	},
	{
		Name: "app_organization_invitationmanager",
		Permissions: []string{
			"app_organization_invitationcreate",
			"app_organization_invitationlist",
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
		Name: RoleProjectOwner,
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
	// group
	{
		Name: GroupOwnerRole,
		Permissions: []string{
			"app_group_administer",
		},
	},
	{
		Name: GroupMemberRole,
		Permissions: []string{
			"app_group_get",
		},
	},
}
