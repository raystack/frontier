package namespace

var systemIdsDefinition = []string{DefinitionTeam.ID, DefinitionUser.ID, DefinitionOrg.ID, DefinitionProject.ID}

var DefinitionOrg = Namespace{
	ID:   "shield/organization",
	Name: "Organization",
}

var DefinitionProject = Namespace{
	ID:   "shield/project",
	Name: "Project",
}

var DefinitionTeam = Namespace{
	ID:   "shield/group",
	Name: "Group",
}

var DefinitionUser = Namespace{
	ID:   "shield/user",
	Name: "User",
}
