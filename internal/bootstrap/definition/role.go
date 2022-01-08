package definition

import (
	"fmt"

	"github.com/odpf/shield/model"
)

var (
	UserType       = UserNamespace.Id
	TeamMemberType = fmt.Sprintf("%s#%s", TeamNamespace.Id, TeamMemberRole.Id)
)

var OrganizationAdminRole = model.Role{
	Name:        "Organization Admin",
	Id:          "organization_admin",
	NamespaceId: OrgNamespace.Id,
	Types:       []string{UserType, TeamMemberType},
}

var ProjectAdminRole = model.Role{
	Name:        "Project Admin",
	Id:          "project_admin",
	NamespaceId: ProjectNamespace.Id,
	Types:       []string{UserType, TeamMemberType},
}

var TeamAdminRole = model.Role{
	Name:        "Team Admin",
	Id:          "team_admin",
	NamespaceId: TeamNamespace.Id,
	Types:       []string{UserType},
}

var TeamMemberRole = model.Role{
	Name:        "Team Member",
	Id:          "team_member",
	NamespaceId: TeamNamespace.Id,
	Types:       []string{UserType},
}
