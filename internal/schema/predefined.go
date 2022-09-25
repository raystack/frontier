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
			SpiceDBPermissionInheritanceFormatter(OrganizationNamespace, OwnerRole),
			SpiceDBPermissionInheritanceFormatter(OrganizationNamespace, EditorRole),
		},
		ViewPermission: {
			OwnerRole, EditorRole, ViewerRole,
			SpiceDBPermissionInheritanceFormatter(OrganizationNamespace, OwnerRole),
			SpiceDBPermissionInheritanceFormatter(OrganizationNamespace, EditorRole),
			SpiceDBPermissionInheritanceFormatter(OrganizationNamespace, ViewerRole),
		},
		DeletePermission: {
			OwnerRole,
			SpiceDBPermissionInheritanceFormatter(OrganizationNamespace, OwnerRole),
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
			SpiceDBPermissionInheritanceFormatter(OrganizationNamespace, OwnerRole),
			SpiceDBPermissionInheritanceFormatter(OrganizationNamespace, EditorRole),
		},
		ViewPermission: {
			ManagerRole, MemberRole,
			SpiceDBPermissionInheritanceFormatter(OrganizationNamespace, OwnerRole),
			SpiceDBPermissionInheritanceFormatter(OrganizationNamespace, EditorRole),
			SpiceDBPermissionInheritanceFormatter(OrganizationNamespace, ViewerRole),
		},
		DeletePermission: {
			ManagerRole,
			SpiceDBPermissionInheritanceFormatter(OrganizationNamespace, OwnerRole),
		},
		MembershipPermission: {
			MemberRole, ManagerRole,
		},
	},
}

var PreDefinedNamespaceConfig = NamespaceConfigMapType{
	UserPrincipal:         NamespaceConfig{},
	OrganizationNamespace: OrganizationNamespaceConfig,
	ProjectNamespace:      ProjectNamespaceConfig,
	TeamNamespace:         TeamNamespaceConfig,
}
