package bootstrap

import (
	"context"
	"fmt"
	"strings"

	"github.com/goto/salt/log"
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
	ns := getResourceNamespace(resource)

	resourceAllActions := getResourceAction("all_actions", ns)

	ownerRole := GetOwnerRole(ns)
	resourceRoles := []model.Role{
		ownerRole,
	}

	actions := []model.Action{
		resourceAllActions,
	}

	policies := getResourceDefaultPolicies(ns, resourceAllActions, ownerRole)

	for action, rolesList := range resource.Actions {
		act := getResourceAction(action, ns)
		actions = append(actions, act)

		for _, r := range rolesList {
			role := getResourceRole(r, ns)
			policy := model.Policy{
				Action:    act,
				Namespace: ns,
				Role:      role,
			}
			resourceRoles = append(resourceRoles, role)
			policies = append(policies, policy)
		}
	}

	s.createNamespaces(ctx, []model.Namespace{ns})
	s.createRoles(ctx, resourceRoles)
	s.createActions(ctx, actions)
	s.createPolicies(ctx, policies)
}

func getResourceDefaultPolicies(ns model.Namespace, action model.Action, owner model.Role) []model.Policy {
	return []model.Policy{
		{
			Action:    action,
			Namespace: ns,
			Role:      definition.TeamAdminRole,
		},
		{
			Action:    action,
			Namespace: ns,
			Role:      definition.ProjectAdminRole,
		},
		{
			Action:    action,
			Namespace: ns,
			Role:      definition.OrganizationAdminRole,
		},
		// {
		//	Action:    action,
		//	Namespace: ns,
		//	Role:      owner,
		// },
	}
}

func GetOwnerRole(ns model.Namespace) model.Role {
	id := fmt.Sprintf("%s_%s", ns.Id, "owner")
	name := fmt.Sprintf("%s_%s", strings.Title(ns.Id), "Owner")
	return model.Role{
		Id:        id,
		Name:      name,
		Types:     []string{definition.UserType},
		Namespace: ns,
	}
}

func getResourceRole(r string, ns model.Namespace) model.Role {
	roleNs := ns
	roleId := fmt.Sprintf("%s_%s", ns.Id, r)

	rSlice := strings.Split(r, ".")

	if len(rSlice) == 2 {
		roleNs = model.Namespace{Id: rSlice[0]}
		roleId = rSlice[1]
	}

	if roleNs.Id == definition.TeamNamespace.Id {
		return model.Role{
			Id:        roleId,
			Name:      roleId,
			Namespace: definition.TeamNamespace,
			Types:     []string{definition.UserType},
		}
	}

	role := model.Role{
		Id:        roleId,
		Name:      roleId,
		Namespace: roleNs,
		Types:     []string{definition.UserType, definition.TeamMemberType},
	}
	return role
}

func getResourceAction(action string, ns model.Namespace) model.Action {
	actId := fmt.Sprintf("%s_%s", ns.Id, action)
	actionName := fmt.Sprintf("%s %s", strings.Title(strings.ToLower(ns.Id)), strings.Title(strings.ToLower(action)))
	act := model.Action{
		Id:        actId,
		Name:      actionName,
		Namespace: ns,
	}
	return act
}

func getResourceNamespace(resource structs.Resource) model.Namespace {
	nsID := utils.Slugify(resource.Name, utils.SlugifyOptions{})
	ns := model.Namespace{
		Name: resource.Name,
		Id:   nsID,
	}
	return ns
}

func (s Service) BootstrapResources(ctx context.Context, resourceConfig *blobstore.ResourcesRepository) error {
	resources, err := resourceConfig.GetAll(ctx)
	if err != nil {
		return err
	}
	for _, resource := range resources {
		s.onboardResource(ctx, resource)
	}

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
		definition.ManageTeamOrgAdminPolicy,
		definition.ManageProjectPolicy,
		definition.ManageProjectOrgPolicy,
		definition.ProjectOrgAdminPolicy,
	}

	s.createPolicies(ctx, policies)

	s.Logger.Info("Bootstrap Polices Successfully")
}

func (s Service) createPolicies(ctx context.Context, policies []model.Policy) {
	for _, policy := range policies {
		_, err := s.SchemaService.CreatePolicy(ctx, policy)
		if err != nil {
			s.Logger.Fatal(err.Error())
		}
	}
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

	s.createActions(ctx, actions)

	s.Logger.Info("Bootstrap Actions Successfully")
}

func (s Service) createActions(ctx context.Context, actions []model.Action) {
	for _, action := range actions {
		_, err := s.SchemaService.CreateAction(ctx, action)
		if err != nil {
			s.Logger.Fatal(err.Error())
		}
	}
}

func (s Service) bootstrapRoles(ctx context.Context) {
	rolesList := []model.Role{
		definition.OrganizationAdminRole,
		definition.ProjectAdminRole,
		definition.TeamAdminRole,
		definition.TeamMemberRole,
		definition.ProjectAdminRole,
	}
	s.createRoles(ctx, rolesList)
	s.Logger.Info("Bootstrap Roles Successfully")
}

func (s Service) createRoles(ctx context.Context, roles []model.Role) {
	for _, role := range roles {
		_, err := s.RoleService.Create(ctx, role)
		if err != nil {
			s.Logger.Fatal(err.Error())
		}
	}
}

func (s Service) createNamespaces(ctx context.Context, namespaces []model.Namespace) {
	for _, ns := range namespaces {
		_, err := s.SchemaService.CreateNamespace(ctx, ns)
		if err != nil {
			s.Logger.Fatal(err.Error())
		}
	}
}

func (s Service) bootstrapNamespaces(ctx context.Context) {
	namespaces := []model.Namespace{
		definition.OrgNamespace,
		definition.ProjectNamespace,
		definition.TeamNamespace,
		definition.UserNamespace,
	}

	s.createNamespaces(ctx, namespaces)
	s.Logger.Info("Bootstrap Namespaces Successfully")
}
