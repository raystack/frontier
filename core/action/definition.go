package action

import "github.com/goto/shield/core/namespace"

var DefinitionManageOrganization = Action{
	ID:          "manage_organization",
	Name:        "Manage Organization",
	NamespaceID: namespace.DefinitionOrg.ID,
}

var DefinitionCreateProject = Action{
	ID:          "create_project",
	Name:        "Create Project",
	NamespaceID: namespace.DefinitionOrg.ID,
}

var DefinitionCreateTeam = Action{
	ID:          "create_team",
	Name:        "Create Team",
	NamespaceID: namespace.DefinitionOrg.ID,
}

var DefinitionManageTeam = Action{
	ID:          "manage_team",
	Name:        "Manage Team",
	NamespaceID: namespace.DefinitionTeam.ID,
}

var DefinitionViewTeam = Action{
	ID:          "view_team",
	Name:        "View Team",
	NamespaceID: namespace.DefinitionTeam.ID,
}

var DefinitionManageProject = Action{
	ID:          "manage_project",
	Name:        "Manage Project",
	NamespaceID: namespace.DefinitionProject.ID,
}

var DefinitionTeamAll = Action{
	ID:          "all_actions_team",
	Name:        "All Actions Team",
	NamespaceID: namespace.DefinitionTeam.ID,
}

var DefinitionProjectAll = Action{
	ID:          "all_actions_project",
	Name:        "All Actions Project",
	NamespaceID: namespace.DefinitionProject.ID,
}
