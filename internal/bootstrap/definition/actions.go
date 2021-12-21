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

var ManageTeamAction = model.Action{
	Id:          "manage_team",
	Name:        "Manage Team",
	NamespaceId: TeamNamespace.Id,
}

var ViewTeamAction = model.Action{
	Id:          "view_team",
	Name:        "View Team",
	NamespaceId: TeamNamespace.Id,
}

var ManageProjectAction = model.Action{
	Id:          "manage_project",
	Name:        "Manage Project",
	NamespaceId: ProjectNamespace.Id,
}

var TeamAllAction = model.Action{
	Id:          "all_actions_team",
	Name:        "All Actions Team",
	NamespaceId: TeamNamespace.Id,
}

var ProjectAllAction = model.Action{
	Id:          "all_actions_project",
	Name:        "All Actions Project",
	NamespaceId: ProjectNamespace.Id,
}
