package namespace

var systemIdsDefinition = []string{DefinitionTeam.Id, DefinitionUser.Id, DefinitionOrg.Id, DefinitionProject.Id}

var DefinitionOrg = Namespace{
	Id:   "organization",
	Name: "Organization",
}

var DefinitionProject = Namespace{
	Id:   "project",
	Name: "Project",
}

var DefinitionTeam = Namespace{
	Id:   "team",
	Name: "Team",
}

var DefinitionUser = Namespace{
	Id:   "user",
	Name: "User",
}
