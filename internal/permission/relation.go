package permission

import (
	"context"

	"github.com/odpf/shield/internal/authz"
	"github.com/odpf/shield/internal/bootstrap"
	"github.com/odpf/shield/internal/bootstrap/definition"
	"github.com/odpf/shield/model"
	"github.com/odpf/shield/pkg/utils"
	blobstore "github.com/odpf/shield/store/blob"
)

type Store interface {
	GetCurrentUser(ctx context.Context, email string) (model.User, error)
	CreateRelation(ctx context.Context, relation model.Relation) (model.Relation, error)
}

type Service struct {
	Authz               *authz.Authz
	Store               Store
	IdentityProxyHeader string
	ResourcesRepository *blobstore.ResourcesRepository
}

type Permissions interface {
	AddTeamToOrg(ctx context.Context, team model.Group, org model.Organization) error
	AddAdminToTeam(ctx context.Context, user model.User, team model.Group) error
	AddMemberToTeam(ctx context.Context, user model.User, team model.Group) error
	AddAdminToOrg(ctx context.Context, user model.User, org model.Organization) error
	AddAdminToProject(ctx context.Context, user model.User, project model.Project) error
	AddProjectToOrg(ctx context.Context, project model.Project, org model.Organization) error
	AddTeamToResource(ctx context.Context, team model.Group, resource model.Resource) error
	AddOwnerToResource(ctx context.Context, user model.User, resource model.Resource) error
	AddProjectToResource(ctx context.Context, project model.Project, resource model.Resource) error
	AddOrgToResource(ctx context.Context, org model.Organization, resource model.Resource) error
	FetchCurrentUser(ctx context.Context) (model.User, error)
	CheckPermission(ctx context.Context, user model.User, resource model.Resource, action model.Action) (bool, error)
}

func (s Service) addRelation(ctx context.Context, rel model.Relation) error {
	newRel, err := s.Store.CreateRelation(ctx, rel)
	if err != nil {
		return err
	}

	err = s.Authz.Permission.AddRelation(ctx, newRel)
	if err != nil {
		return err
	}
	return nil
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
		RelationType: model.RelationTypes.Namespace,
	}

	return s.addRelation(ctx, rel)
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
	return s.addRelation(ctx, rel)
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
	return s.addRelation(ctx, rel)
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
	return s.addRelation(ctx, rel)
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
	return s.addRelation(ctx, rel)
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
		RelationType: model.RelationTypes.Namespace,
	}
	return s.addRelation(ctx, rel)

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
		RelationType: model.RelationTypes.Namespace,
	}
	return s.addRelation(ctx, rel)

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
		RelationType: model.RelationTypes.Namespace,
	}
	return s.addRelation(ctx, rel)

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
		RelationType: model.RelationTypes.Namespace,
	}
	return s.addRelation(ctx, rel)

}

func (s Service) CheckPermission(ctx context.Context, user model.User, resource model.Resource, action model.Action) (bool, error) {
	resourceNS := model.Namespace{
		Id: utils.DefaultStringIfEmpty(resource.NamespaceId, resource.Namespace.Id),
	}

	rel := model.Relation{
		ObjectNamespace:  resourceNS,
		ObjectId:         resource.Id,
		SubjectId:        user.Id,
		SubjectNamespace: definition.UserNamespace,
	}

	return s.Authz.Permission.CheckRelation(ctx, rel, action)
}

func (s Service) AddOwnerToResource(ctx context.Context, user model.User, resource model.Resource) error {
	resourceNS := model.Namespace{
		Id: resource.NamespaceId,
	}

	relationSet, err := s.ResourcesRepository.GetRelationsForNamespace(ctx, resource.NamespaceId)
	if err != nil {
		return err
	}

	role := bootstrap.GetOwnerRole(resourceNS)

	if !relationSet[role.Id] {
		return nil
	}

	rel := model.Relation{
		ObjectNamespace:  resourceNS,
		ObjectId:         resource.Id,
		SubjectId:        user.Id,
		SubjectNamespace: definition.UserNamespace,
		Role:             role,
	}

	return s.addRelation(ctx, rel)
}
