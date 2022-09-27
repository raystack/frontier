package schema

// SpiceDB readable format is stored in predefined_schema.txt
const (
	// namespace
	OrganizationNamespace = "organization"
	ProjectNamespace      = "project"
	TeamNamespace         = "team"

	// roles
	OwnerRole   = "owner"
	EditorRole  = "editor"
	ViewerRole  = "viewer"
	ManagerRole = "manager"
	MemberRole  = "member"

	// permissions
	ViewPermission   = "view"
	EditPermission   = "edit"
	DeletePermission = "delete"

	// synthetic permission
	MembershipPermission = "membership"

	// principals
	UserPrincipal = "user"
	TeamPrincipal = "team"
)

var OrganizationNamespaceConfig = NamespaceConfig{
	Roles: map[string][]string{
		OwnerRole:  {UserPrincipal, TeamPrincipal},
		EditorRole: {UserPrincipal, TeamPrincipal},
		ViewerRole: {UserPrincipal, TeamPrincipal},
	},
	Permissions: map[string][]string{
		EditPermission: {
			OwnerRole, EditorRole,
		},
		ViewPermission: {
			OwnerRole, EditorRole, ViewerRole,
		},
	},
}

var ProjectNamespaceConfig = NamespaceConfig{
	InheritedNamespaces: []string{OrganizationNamespace},
	Roles: map[string][]string{
		OwnerRole:  {UserPrincipal, TeamPrincipal},
		EditorRole: {UserPrincipal, TeamPrincipal},
		ViewerRole: {UserPrincipal, TeamPrincipal},
	},
	Permissions: map[string][]string{
		EditPermission: {
			OwnerRole, EditorRole,
			PermissionInheritanceFormatter(OrganizationNamespace, OwnerRole),
			PermissionInheritanceFormatter(OrganizationNamespace, EditorRole),
		},
		ViewPermission: {
			OwnerRole, EditorRole, ViewerRole,
			PermissionInheritanceFormatter(OrganizationNamespace, OwnerRole),
			PermissionInheritanceFormatter(OrganizationNamespace, EditorRole),
			PermissionInheritanceFormatter(OrganizationNamespace, ViewerRole),
		},
		DeletePermission: {
			OwnerRole,
			PermissionInheritanceFormatter(OrganizationNamespace, OwnerRole),
		},
	},
}

var TeamNamespaceConfig = NamespaceConfig{
	InheritedNamespaces: []string{OrganizationNamespace},
	Roles: map[string][]string{
		MemberRole:  {UserPrincipal},
		ManagerRole: {UserPrincipal},
	},
	Permissions: map[string][]string{
		EditPermission: {
			ManagerRole,
			PermissionInheritanceFormatter(OrganizationNamespace, OwnerRole),
			PermissionInheritanceFormatter(OrganizationNamespace, EditorRole),
		},
		ViewPermission: {
			ManagerRole, MemberRole,
			PermissionInheritanceFormatter(OrganizationNamespace, OwnerRole),
			PermissionInheritanceFormatter(OrganizationNamespace, EditorRole),
			PermissionInheritanceFormatter(OrganizationNamespace, ViewerRole),
		},
		DeletePermission: {
			ManagerRole,
			PermissionInheritanceFormatter(OrganizationNamespace, OwnerRole),
		},
		MembershipPermission: {
			MemberRole, ManagerRole,
		},
	},
}

var PreDefinedSystemNamespaceConfig = NamespaceConfigMapType{
	UserPrincipal:         NamespaceConfig{},
	OrganizationNamespace: OrganizationNamespaceConfig,
	ProjectNamespace:      ProjectNamespaceConfig,
	TeamNamespace:         TeamNamespaceConfig,
}

var PreDefinedResourceGroupNamespaceConfig = NamespaceConfig{
	Type:                ResourceGroupNamespace,
	InheritedNamespaces: []string{OrganizationNamespace, ProjectNamespace},
	Roles: map[string][]string{
		OwnerRole:  {UserPrincipal, TeamPrincipal},
		EditorRole: {UserPrincipal, TeamPrincipal},
		ViewerRole: {UserPrincipal, TeamPrincipal},
	},
	Permissions: map[string][]string{
		EditPermission: {
			OwnerRole, EditorRole,
			PermissionInheritanceFormatter(OrganizationNamespace, OwnerRole),
			PermissionInheritanceFormatter(OrganizationNamespace, EditorRole),
			PermissionInheritanceFormatter(ProjectNamespace, OwnerRole),
			PermissionInheritanceFormatter(ProjectNamespace, EditorRole),
		},
		ViewPermission: {
			OwnerRole, EditorRole, ViewerRole,
			PermissionInheritanceFormatter(OrganizationNamespace, OwnerRole),
			PermissionInheritanceFormatter(OrganizationNamespace, EditorRole),
			PermissionInheritanceFormatter(OrganizationNamespace, ViewerRole),
			PermissionInheritanceFormatter(ProjectNamespace, OwnerRole),
			PermissionInheritanceFormatter(ProjectNamespace, EditorRole),
			PermissionInheritanceFormatter(ProjectNamespace, ViewerRole),
		},
		DeletePermission: {
			OwnerRole,
			PermissionInheritanceFormatter(OrganizationNamespace, OwnerRole),
			PermissionInheritanceFormatter(ProjectNamespace, OwnerRole),
		},
	},
}
