package action

import "github.com/odpf/shield/core/namespace"

var DefinitionManageOrganization = Action{
	Id:          "manage_organization",
	Name:        "Manage Organization",
	NamespaceId: namespace.DefinitionOrg.Id,
}

var DefinitionCreateProject = Action{
	Id:          "create_project",
	Name:        "Create Project",
	NamespaceId: namespace.DefinitionOrg.Id,
}

var DefinitionCreateTeam = Action{
	Id:          "create_team",
	Name:        "Create Team",
	NamespaceId: namespace.DefinitionOrg.Id,
}

var DefinitionManageTeam = Action{
	Id:          "manage_team",
	Name:        "Manage Team",
	NamespaceId: namespace.DefinitionTeam.Id,
}

var DefinitionViewTeam = Action{
	Id:          "view_team",
	Name:        "View Team",
	NamespaceId: namespace.DefinitionTeam.Id,
}

var DefinitionManageProject = Action{
	Id:          "manage_project",
	Name:        "Manage Project",
	NamespaceId: namespace.DefinitionProject.Id,
}

var DefinitionTeamAll = Action{
	Id:          "all_actions_team",
	Name:        "All Actions Team",
	NamespaceId: namespace.DefinitionTeam.Id,
}

var DefinitionProjectAll = Action{
	Id:          "all_actions_project",
	Name:        "All Actions Project",
	NamespaceId: namespace.DefinitionProject.Id,
}
