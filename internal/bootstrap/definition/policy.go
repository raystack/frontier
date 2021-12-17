package definition

import "github.com/odpf/shield/model"

var OrganizationManagePolicy = model.Policy{
	NamespaceId: OrgNamespace.Id,
	RoleId:      OrganizationAdminRole.Id,
	ActionId:    ManageOrganizationAction.Id,
}

var CreateProjectPolicy = model.Policy{
	NamespaceId: OrgNamespace.Id,
	RoleId:      OrganizationAdminRole.Id,
	ActionId:    CreateProjectAction.Id,
}

var CreateTeamPolicy = model.Policy{
	NamespaceId: OrgNamespace.Id,
	RoleId:      OrganizationAdminRole.Id,
	ActionId:    CreateTeamAction.Id,
}
