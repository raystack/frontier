package definition

import "github.com/odpf/shield/model"

var ManageOrganizationAction = model.Action{
	Id:          "manage_organization",
	Name:        "Manage Organization",
	NamespaceId: OrgNamespace.Id,
}

var CreateProjectAction = model.Action{
	Id:          "create_project",
	Name:        "Create Project",
	NamespaceId: OrgNamespace.Id,
}

var CreateTeamAction = model.Action{
	Id:          "create_team",
	Name:        "Create Team",
	NamespaceId: OrgNamespace.Id,
}
