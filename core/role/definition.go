package role

import (
	"fmt"

	"github.com/raystack/shield/core/namespace"
)

var (
	UserType       = namespace.DefinitionUser.ID
	TeamMemberType = fmt.Sprintf("%s#%s", namespace.DefinitionTeam.ID, DefinitionTeamMember.ID)
)

var DefinitionOrganizationAdmin = Role{
	Name:        "Organization Admin",
	ID:          "organization_admin",
	NamespaceID: namespace.DefinitionOrg.ID,
	Types:       []string{UserType, TeamMemberType},
}

var DefinitionProjectAdmin = Role{
	Name:        "Project Admin",
	ID:          "project_admin",
	NamespaceID: namespace.DefinitionProject.ID,
	Types:       []string{UserType, TeamMemberType},
}

var DefinitionTeamAdmin = Role{
	Name:        "Team Admin",
	ID:          "team_admin",
	NamespaceID: namespace.DefinitionTeam.ID,
	Types:       []string{UserType},
}

var DefinitionTeamMember = Role{
	Name:        "Team Member",
	ID:          "team_member",
	NamespaceID: namespace.DefinitionTeam.ID,
	Types:       []string{UserType},
}
