package bootstrap

import (
	"context"
	"fmt"

	"github.com/odpf/salt/log"
	"github.com/odpf/shield/internal/bootstrap/definition"
	"github.com/odpf/shield/internal/roles"
	"github.com/odpf/shield/internal/schema"
	"github.com/odpf/shield/model"
	"github.com/odpf/shield/pkg/utils"
	blobstore "github.com/odpf/shield/store/blob"
	"github.com/odpf/shield/structs"
)

// Insert Action
// Insert Policy

type Service struct {
	Logger        log.Logger
	SchemaService schema.Service
	RoleService   roles.Service
}

func (s Service) BootstrapDefaultDefinitions(ctx context.Context) {
	s.bootstrapNamespaces(ctx)
	s.bootstrapRoles(ctx)
	s.bootstrapActions(ctx)
	s.bootstrapPolicies(ctx)
}

func (s Service) onboardResource(ctx context.Context, resource structs.Resource) {
	nsID := utils.Slugify(resource.Name, utils.SlugifyOptions{})
	ns := model.Namespace{
		Name: resource.Name,
		Id:   nsID,
	}
	_, err := s.SchemaService.CreateNamespace(ctx, ns)
	if err != nil {
		s.Logger.Fatal(err.Error())
	}

	resourceRoles := []model.Role{}
	actions := []model.Action{}
	policies := []model.Policy{}

	for action, roles := range resource.Actions {
		actId := fmt.Sprintf("%s_%s", nsID, action)
		act := model.Action{
			Id:        actId,
			Name:      action,
			Namespace: ns,
		}
		actions = append(actions, act)

		for _, r := range roles {
			roleId := fmt.Sprintf("%s_%s", nsID, r)
			role := model.Role{
				Id:        roleId,
				Name:      roleId,
				Namespace: ns,
				Types:     []string{definition.UserType, definition.TeamMemberType},
			}
			resourceRoles = append(resourceRoles, role)

			policy := model.Policy{
				Action:    act,
				Namespace: ns,
				Role:      role,
			}

			policies = append(policies, policy)
		}
	}

	for _, role := range resourceRoles {
		_, err := s.RoleService.Create(ctx, role)
		if err != nil {
			s.Logger.Fatal(err.Error())
		}
	}

	for _, act := range actions {
		_, err := s.SchemaService.CreateAction(ctx, act)
		if err != nil {
			s.Logger.Fatal(err.Error())
		}
	}

	for _, p := range policies {
		_, err := s.SchemaService.CreatePolicy(ctx, p)
		if err != nil {
			s.Logger.Fatal(err.Error())
		}
	}
}

func (s Service) BootstrapResources(ctx context.Context, resourceConfig *blobstore.ResourcesRepository) error {
	resources, err := resourceConfig.GetAll(ctx)
	if err != nil {
		return err
	}
	for _, resource := range resources {
		s.onboardResource(ctx, resource)
	}

	// Onboard resource
	// Create resource roles
	// Create actions
	// Create Policy
	return nil
}

func (s Service) bootstrapPolicies(ctx context.Context) {
	policies := []model.Policy{
		definition.ViewTeamMemberPolicy,
		definition.ViewTeamAdminPolicy,
		definition.OrganizationManagePolicy,
		definition.CreateProjectPolicy,
		definition.CreateTeamPolicy,
		definition.ManageTeamPolicy,
		definition.TeamOrgAdminPolicy,
		definition.ManageProjectPolicy,
		definition.ManageProjectOrgPolicy,
		definition.ProjectOrgAdminPolicy,
	}

	for _, policy := range policies {
		_, err := s.SchemaService.CreatePolicy(ctx, policy)
		if err != nil {
			s.Logger.Fatal(err.Error())
		}
	}

	s.Logger.Info("Bootstrap Polices Successfully")
}

func (s Service) bootstrapActions(ctx context.Context) {
	actions := []model.Action{
		definition.ManageOrganizationAction,
		definition.CreateProjectAction,
		definition.CreateTeamAction,
		definition.ManageTeamAction,
		definition.ViewTeamAction,
		definition.ManageProjectAction,
		definition.TeamAllAction,
		definition.ProjectAllAction,
	}

	for _, action := range actions {
		_, err := s.SchemaService.CreateAction(ctx, action)
		if err != nil {
			s.Logger.Fatal(err.Error())
		}
	}

	s.Logger.Info("Bootstrap Actions Successfully")
}

func (s Service) bootstrapRoles(ctx context.Context) {
	roles := []model.Role{
		definition.OrganizationAdminRole,
		definition.ProjectAdminRole,
		definition.TeamAdminRole,
		definition.TeamMemberRole,
		definition.ProjectAdminRole,
	}

	for _, role := range roles {
		_, err := s.RoleService.Create(ctx, role)
		if err != nil {
			s.Logger.Fatal(err.Error())
		}
	}

	s.Logger.Info("Bootstrap Roles Successfully")
}

func (s Service) bootstrapNamespaces(ctx context.Context) {
	namespaces := []model.Namespace{
		definition.OrgNamespace,
		definition.ProjectNamespace,
		definition.TeamNamespace,
		definition.UserNamespace,
	}

	for _, ns := range namespaces {
		_, err := s.SchemaService.CreateNamespace(ctx, ns)
		if err != nil {
			s.Logger.Fatal(err.Error())
		}
	}
	s.Logger.Info("Bootstrap Namespaces Successfully")
}
