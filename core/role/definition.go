package role

import (
	"fmt"

	"github.com/odpf/shield/core/namespace"
)

var (
	UserType       = namespace.DefinitionUser.Id
	TeamMemberType = fmt.Sprintf("%s#%s", namespace.DefinitionTeam.Id, DefinitionTeamMember.Id)
)

var DefinitionOrganizationAdmin = Role{
	Name:        "Organization Admin",
	Id:          "organization_admin",
	NamespaceId: namespace.DefinitionOrg.Id,
	Types:       []string{UserType, TeamMemberType},
}

var DefinitionProjectAdmin = Role{
	Name:        "Project Admin",
	Id:          "project_admin",
	NamespaceId: namespace.DefinitionProject.Id,
	Types:       []string{UserType, TeamMemberType},
}

var DefinitionTeamAdmin = Role{
	Name:        "Team Admin",
	Id:          "team_admin",
	NamespaceId: namespace.DefinitionTeam.Id,
	Types:       []string{UserType},
}

var DefinitionTeamMember = Role{
	Name:        "Team Member",
	Id:          "team_member",
	NamespaceId: namespace.DefinitionTeam.Id,
	Types:       []string{UserType},
}
