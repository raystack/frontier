package namespace

var systemIdsDefinition = []string{DefinitionTeam.ID, DefinitionUser.ID, DefinitionOrg.ID, DefinitionProject.ID}

var DefinitionOrg = Namespace{
	ID:   "organization",
	Name: "Organization",
}

var DefinitionProject = Namespace{
	ID:   "project",
	Name: "Project",
}

var DefinitionTeam = Namespace{
	ID:   "group",
	Name: "Group",
}

var DefinitionUser = Namespace{
	ID:   "user",
	Name: "User",
}
