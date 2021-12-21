package bootstrap

import (
	"context"

	"github.com/odpf/salt/log"
	"github.com/odpf/shield/internal/bootstrap/definition"
	"github.com/odpf/shield/internal/roles"
	"github.com/odpf/shield/internal/schema"
	"github.com/odpf/shield/model"
)

// Insert Action
// Insert Policy

type Service struct {
	Logger        log.Logger
	SchemaService schema.Service
	RoleService   roles.Service
}

func (s Service) BootstrapDefinitions(ctx context.Context) {
	s.bootstrapNamespaces(ctx)
	s.bootstrapRoles(ctx)
	s.bootstrapActions(ctx)
	s.bootstrapPolicies(ctx)
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
