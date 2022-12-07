package schema

// SpiceDB readable format is stored in predefined_schema.txt
const (
	// namespace
	OrganizationNamespace = "shield/organization"
	ProjectNamespace      = "shield/project"
	GroupNamespace        = "shield/group"

	// relation
	OrganizationRelationName = "organization"
	ProjectRelationName      = "project"
	GroupRelationName        = "group"

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
	UserPrincipal  = "shield/user"
	GroupPrincipal = "shield/group"
)

var InheritedRelations = map[string]bool{
	OrganizationRelationName: true,
	ProjectRelationName:      true,
}

var OrganizationNamespaceConfig = NamespaceConfig{
	Roles: map[string][]string{
		OwnerRole:  {UserPrincipal, GroupPrincipal},
		EditorRole: {UserPrincipal, GroupPrincipal},
		ViewerRole: {UserPrincipal, GroupPrincipal},
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
	InheritedNamespaces: []InheritedNamespace{
		{
			Name:        OrganizationRelationName,
			NamespaceId: OrganizationNamespace,
		},
	},
	Roles: map[string][]string{
		OwnerRole:  {UserPrincipal, GroupPrincipal},
		EditorRole: {UserPrincipal, GroupPrincipal},
		ViewerRole: {UserPrincipal, GroupPrincipal},
	},
	Permissions: map[string][]string{
		EditPermission: {
			OwnerRole, EditorRole,
			PermissionInheritanceFormatter(OrganizationRelationName, OwnerRole),
			PermissionInheritanceFormatter(OrganizationRelationName, EditorRole),
		},
		ViewPermission: {
			OwnerRole, EditorRole, ViewerRole,
			PermissionInheritanceFormatter(OrganizationRelationName, OwnerRole),
			PermissionInheritanceFormatter(OrganizationRelationName, EditorRole),
			PermissionInheritanceFormatter(OrganizationRelationName, ViewerRole),
		},
		DeletePermission: {
			OwnerRole,
			PermissionInheritanceFormatter(OrganizationRelationName, OwnerRole),
		},
	},
}

var GroupNamespaceConfig = NamespaceConfig{
	InheritedNamespaces: []InheritedNamespace{
		{
			Name:        OrganizationRelationName,
			NamespaceId: OrganizationNamespace,
		},
	},
	Roles: map[string][]string{
		MemberRole:  {UserPrincipal},
		ManagerRole: {UserPrincipal},
	},
	Permissions: map[string][]string{
		EditPermission: {
			ManagerRole,
			PermissionInheritanceFormatter(OrganizationRelationName, OwnerRole),
			PermissionInheritanceFormatter(OrganizationRelationName, EditorRole),
		},
		ViewPermission: {
			ManagerRole, MemberRole,
			PermissionInheritanceFormatter(OrganizationRelationName, OwnerRole),
			PermissionInheritanceFormatter(OrganizationRelationName, EditorRole),
			PermissionInheritanceFormatter(OrganizationRelationName, ViewerRole),
		},
		DeletePermission: {
			ManagerRole,
			PermissionInheritanceFormatter(OrganizationRelationName, OwnerRole),
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
	GroupNamespace:        GroupNamespaceConfig,
}

var PreDefinedResourceGroupNamespaceConfig = NamespaceConfig{
	Type: ResourceGroupNamespace,
	InheritedNamespaces: []InheritedNamespace{
		{
			Name:        OrganizationRelationName,
			NamespaceId: OrganizationNamespace,
		},
		{
			Name:        ProjectRelationName,
			NamespaceId: ProjectNamespace,
		},
	},
	Roles: map[string][]string{
		OwnerRole:  {UserPrincipal, GroupPrincipal},
		EditorRole: {UserPrincipal, GroupPrincipal},
		ViewerRole: {UserPrincipal, GroupPrincipal},
	},
	Permissions: map[string][]string{
		EditPermission: {
			OwnerRole, EditorRole,
			PermissionInheritanceFormatter(OrganizationRelationName, OwnerRole),
			PermissionInheritanceFormatter(OrganizationRelationName, EditorRole),
			PermissionInheritanceFormatter(ProjectRelationName, OwnerRole),
			PermissionInheritanceFormatter(ProjectRelationName, EditorRole),
		},
		ViewPermission: {
			OwnerRole, EditorRole, ViewerRole,
			PermissionInheritanceFormatter(OrganizationRelationName, OwnerRole),
			PermissionInheritanceFormatter(OrganizationRelationName, EditorRole),
			PermissionInheritanceFormatter(OrganizationRelationName, ViewerRole),
			PermissionInheritanceFormatter(ProjectRelationName, OwnerRole),
			PermissionInheritanceFormatter(ProjectRelationName, EditorRole),
			PermissionInheritanceFormatter(ProjectRelationName, ViewerRole),
		},
		DeletePermission: {
			OwnerRole,
			PermissionInheritanceFormatter(OrganizationRelationName, OwnerRole),
			PermissionInheritanceFormatter(ProjectRelationName, OwnerRole),
		},
	},
}
