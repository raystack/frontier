package permission

import (
	"context"

	"github.com/odpf/shield/internal/authz"
	"github.com/odpf/shield/internal/bootstrap/definition"
	"github.com/odpf/shield/model"
	"github.com/odpf/shield/pkg/utils"
)

type Store interface {
	GetCurrentUser(ctx context.Context, email string) (model.User, error)
}

type Service struct {
	Authz               *authz.Authz
	Store               Store
	IdentityProxyHeader string
}

type Permissions interface {
	AddTeamToOrg(ctx context.Context, team model.Group, org model.Organization) error
	AddAdminToTeam(ctx context.Context, user model.User, team model.Group) error
	AddAdminToOrg(ctx context.Context, user model.User, org model.Organization) error
	AddAdminToProject(ctx context.Context, user model.User, project model.Project) error
	AddProjectToOrg(ctx context.Context, project model.Project, org model.Organization) error
	AddTeamToResource(ctx context.Context, team model.Group, resource model.Resource) error
	AddProjectToResource(ctx context.Context, project model.Project, resource model.Resource) error
	AddOrgToResource(ctx context.Context, org model.Organization, resource model.Resource) error
	FetchCurrentUser(ctx context.Context) (model.User, error)
	CheckPermission(ctx context.Context, user model.User, resource model.Resource, permission model.Permission) (bool, error)
}

func (s Service) AddTeamToOrg(ctx context.Context, team model.Group, org model.Organization) error {
	orgId := utils.DefaultStringIfEmpty(org.Id, team.OrganizationId)
	rel := model.Relation{
		ObjectNamespace:  definition.TeamNamespace,
		ObjectId:         team.Id,
		SubjectId:        orgId,
		SubjectNamespace: definition.OrgNamespace,
		Role: model.Role{
			Id:        definition.OrgNamespace.Id,
			Namespace: definition.TeamNamespace,
		},
	}
	err := s.Authz.Permission.AddRelation(ctx, rel)
	if err != nil {
		return err
	}
	return nil
}

func (s Service) AddAdminToTeam(ctx context.Context, user model.User, team model.Group) error {
	rel := model.Relation{
		ObjectNamespace:  definition.TeamNamespace,
		ObjectId:         team.Id,
		SubjectId:        user.Id,
		SubjectNamespace: definition.UserNamespace,
		Role: model.Role{
			Id:        definition.TeamAdminRole.Id,
			Namespace: definition.TeamNamespace,
		},
	}
	err := s.Authz.Permission.AddRelation(ctx, rel)
	if err != nil {
		return err
	}
	return nil
}

func (s Service) AddMemberToTeam(ctx context.Context, user model.User, team model.Group) error {
	rel := model.Relation{
		ObjectNamespace:  definition.TeamNamespace,
		ObjectId:         team.Id,
		SubjectId:        user.Id,
		SubjectNamespace: definition.UserNamespace,
		Role: model.Role{
			Id:        definition.TeamMemberRole.Id,
			Namespace: definition.TeamNamespace,
		},
	}
	err := s.Authz.Permission.AddRelation(ctx, rel)
	if err != nil {
		return err
	}
	return nil
}

func (s Service) AddAdminToOrg(ctx context.Context, user model.User, org model.Organization) error {
	rel := model.Relation{
		ObjectNamespace:  definition.OrgNamespace,
		ObjectId:         org.Id,
		SubjectId:        user.Id,
		SubjectNamespace: definition.UserNamespace,
		Role: model.Role{
			Id:        definition.OrganizationAdminRole.Id,
			Namespace: definition.OrgNamespace,
		},
	}
	err := s.Authz.Permission.AddRelation(ctx, rel)
	if err != nil {
		return err
	}
	return nil
}

func (s Service) AddAdminToProject(ctx context.Context, user model.User, project model.Project) error {
	rel := model.Relation{
		ObjectNamespace:  definition.ProjectNamespace,
		ObjectId:         project.Id,
		SubjectId:        user.Id,
		SubjectNamespace: definition.UserNamespace,
		Role: model.Role{
			Id:        definition.ProjectAdminRole.Id,
			Namespace: definition.ProjectNamespace,
		},
	}
	err := s.Authz.Permission.AddRelation(ctx, rel)
	if err != nil {
		return err
	}
	return nil
}

func (s Service) AddProjectToOrg(ctx context.Context, project model.Project, org model.Organization) error {
	rel := model.Relation{
		ObjectNamespace:  definition.ProjectNamespace,
		ObjectId:         project.Id,
		SubjectId:        org.Id,
		SubjectNamespace: definition.OrgNamespace,
		Role: model.Role{
			Id:        definition.OrgNamespace.Id,
			Namespace: definition.ProjectNamespace,
		},
	}
	err := s.Authz.Permission.AddRelation(ctx, rel)
	if err != nil {
		return err
	}
	return nil
}

func (s Service) AddProjectToResource(ctx context.Context, project model.Project, resource model.Resource) error {
	resourceNS := model.Namespace{
		Id: resource.NamespaceId,
	}

	rel := model.Relation{
		ObjectNamespace:  resourceNS,
		ObjectId:         resource.Id,
		SubjectId:        project.Id,
		SubjectNamespace: definition.ProjectNamespace,
		Role: model.Role{
			Id:        definition.ProjectNamespace.Id,
			Namespace: resourceNS,
		},
	}
	err := s.Authz.Permission.AddRelation(ctx, rel)
	if err != nil {
		return err
	}
	return nil
}

func (s Service) AddOrgToResource(ctx context.Context, org model.Organization, resource model.Resource) error {
	resourceNS := model.Namespace{
		Id: resource.NamespaceId,
	}

	rel := model.Relation{
		ObjectNamespace:  resourceNS,
		ObjectId:         resource.Id,
		SubjectId:        org.Id,
		SubjectNamespace: definition.OrgNamespace,
		Role: model.Role{
			Id:        definition.OrgNamespace.Id,
			Namespace: resourceNS,
		},
	}
	err := s.Authz.Permission.AddRelation(ctx, rel)
	if err != nil {
		return err
	}
	return nil
}

func (s Service) AddTeamToResource(ctx context.Context, team model.Group, resource model.Resource) error {
	resourceNS := model.Namespace{
		Id: resource.NamespaceId,
	}

	rel := model.Relation{
		ObjectNamespace:  resourceNS,
		ObjectId:         resource.Id,
		SubjectId:        team.Id,
		SubjectNamespace: definition.TeamNamespace,
		Role: model.Role{
			Id:        definition.TeamNamespace.Id,
			Namespace: resourceNS,
		},
	}
	err := s.Authz.Permission.AddRelation(ctx, rel)
	if err != nil {
		return err
	}
	return nil
}

func (s Service) CheckPermission(ctx context.Context, user model.User, resource model.Resource, permission model.Permission) (bool, error) {
	resourceNS := model.Namespace{
		Id: resource.NamespaceId,
	}

	rel := model.Relation{
		ObjectNamespace:  resourceNS,
		ObjectId:         resource.Id,
		SubjectId:        user.Id,
		SubjectNamespace: definition.UserNamespace,
	}

	return s.Authz.Permission.CheckRelation(ctx, rel, permission)
}
