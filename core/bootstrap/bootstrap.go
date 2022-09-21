package bootstrap

import (
	"context"
	"fmt"
	"strings"

	"github.com/odpf/salt/log"
	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/policy"
	"github.com/odpf/shield/core/resource"
	"github.com/odpf/shield/core/role"
	"github.com/odpf/shield/pkg/str"
)

type PolicyService interface {
	Create(ctx context.Context, pol policy.Policy) ([]policy.Policy, error)
}

type ActionService interface {
	Create(ctx context.Context, act action.Action) (action.Action, error)
}

type RoleService interface {
	Create(ctx context.Context, rl role.Role) (role.Role, error)
}

type NamespaceService interface {
	Create(ctx context.Context, ns namespace.Namespace) (namespace.Namespace, error)
}

type ResourceService interface {
	GetAllConfigs(ctx context.Context) ([]resource.YAML, error)
}

type Service struct {
	logger           log.Logger
	policyService    PolicyService
	actionService    ActionService
	namespaceService NamespaceService
	roleService      RoleService
	resourceService  ResourceService
}

func NewService(
	logger log.Logger,
	policyService PolicyService,
	actionService ActionService,
	namespaceService NamespaceService,
	roleService RoleService,
	resourceService ResourceService,
) *Service {
	return &Service{
		logger:           logger,
		policyService:    policyService,
		actionService:    actionService,
		namespaceService: namespaceService,
		roleService:      roleService,
		resourceService:  resourceService,
	}
}

func (s Service) BootstrapDefaultDefinitions(ctx context.Context) {
	s.bootstrapNamespaces(ctx)
	s.bootstrapRoles(ctx)
	s.bootstrapActions(ctx)
	s.bootstrapPolicies(ctx)
}

func (s Service) onboardResource(ctx context.Context, resYAML resource.YAML) error {
	ns := getResourceNamespace(resYAML)

	resourceAllActions := getResourceAction("all_actions", ns)

	ownerRole := role.GetOwnerRole(ns)
	resourceRoles := []role.Role{
		ownerRole,
	}

	actions := []action.Action{
		resourceAllActions,
	}

	policies := getResourceDefaultPolicies(ns, resourceAllActions, ownerRole)

	for action, rolesList := range resYAML.Actions {
		act := getResourceAction(action, ns)
		actions = append(actions, act)

		for _, r := range rolesList {
			role := getResourceRole(r, ns)
			policy := policy.Policy{
				Action:      act,
				Namespace:   ns,
				NamespaceID: ns.ID,
				Role:        role,
			}
			resourceRoles = append(resourceRoles, role)
			policies = append(policies, policy)
		}
	}

	if err := s.createNamespaces(ctx, []namespace.Namespace{ns}); err != nil {
		return err
	}
	if err := s.createRoles(ctx, resourceRoles); err != nil {
		return err
	}
	if err := s.createActions(ctx, actions); err != nil {
		return err
	}
	if err := s.createPolicies(ctx, policies); err != nil {
		return err
	}
	return nil
}

func getResourceDefaultPolicies(ns namespace.Namespace, action action.Action, owner role.Role) []policy.Policy {
	return []policy.Policy{
		{
			Action:    action,
			Namespace: ns,
			Role:      role.DefinitionTeamAdmin,
		},
		{
			Action:    action,
			Namespace: ns,
			Role:      role.DefinitionProjectAdmin,
		},
		{
			Action:    action,
			Namespace: ns,
			Role:      role.DefinitionOrganizationAdmin,
		},
		//{
		//	Action:    action,
		//	Namespace: ns,
		//	Role:      owner,
		//},
	}
}

func getResourceRole(r string, ns namespace.Namespace) role.Role {
	roleNs := ns
	roleId := fmt.Sprintf("%s_%s", ns.ID, r)

	rSlice := strings.Split(r, ".")

	if len(rSlice) == 2 {
		roleNs = namespace.Namespace{ID: rSlice[0]}
		roleId = rSlice[1]
	}

	if roleNs.ID == namespace.DefinitionTeam.ID {
		return role.Role{
			ID:          roleId,
			Name:        roleId,
			NamespaceID: namespace.DefinitionTeam.ID,
			Types:       []string{role.UserType},
		}
	}

	role := role.Role{
		ID:          roleId,
		Name:        roleId,
		NamespaceID: roleNs.ID,
		Types:       []string{role.UserType, role.TeamMemberType},
	}
	return role
}

func getResourceAction(actionStr string, ns namespace.Namespace) action.Action {
	actId := fmt.Sprintf("%s_%s", ns.ID, actionStr)
	actionName := fmt.Sprintf("%s %s", strings.Title(strings.ToLower(ns.ID)), strings.Title(strings.ToLower(actionStr)))
	act := action.Action{
		ID:          actId,
		Name:        actionName,
		NamespaceID: ns.ID,
	}
	return act
}

func getResourceNamespace(resYAML resource.YAML) namespace.Namespace {
	nsID := str.Slugify(resYAML.Name, str.SlugifyOptions{})
	ns := namespace.Namespace{
		Name:         resYAML.Name,
		ID:           nsID,
		Backend:      resYAML.Backend,
		ResourceType: resYAML.ResourceType,
	}
	return ns
}

func (s Service) BootstrapResources(ctx context.Context) error {
	resources, err := s.resourceService.GetAllConfigs(ctx)
	if err != nil {
		return err
	}
	for _, resource := range resources {
		if err := s.onboardResource(ctx, resource); err != nil {
			return err
		}
	}

	return nil
}

func (s Service) bootstrapPolicies(ctx context.Context) error {
	policies := []policy.Policy{
		policy.DefinitionViewTeamMember,
		policy.DefinitionViewTeamAdmin,
		policy.DefinitionOrganizationManage,
		policy.DefinitionCreateProject,
		policy.DefinitionCreateTeam,
		policy.DefinitionManageTeam,
		policy.DefinitionTeamOrgAdmin,
		policy.DefinitionManageTeamOrgAdmin,
		policy.DefinitionManageProject,
		policy.DefinitionManageProjectOrg,
		policy.DefinitionProjectOrgAdmin,
	}

	if err := s.createPolicies(ctx, policies); err != nil {
		return err
	}

	s.logger.Info("Bootstrap Polices Successfully")
	return nil
}

func (s Service) createPolicies(ctx context.Context, policies []policy.Policy) error {
	for _, policy := range policies {
		if _, err := s.policyService.Create(ctx, policy); err != nil {
			return err
		}
	}
	return nil
}

func (s Service) bootstrapActions(ctx context.Context) error {
	actions := []action.Action{
		action.DefinitionManageOrganization,
		action.DefinitionCreateProject,
		action.DefinitionCreateTeam,
		action.DefinitionManageTeam,
		action.DefinitionViewTeam,
		action.DefinitionManageProject,
		action.DefinitionTeamAll,
		action.DefinitionProjectAll,
	}

	if err := s.createActions(ctx, actions); err != nil {
		return err
	}

	s.logger.Info("Bootstrap Actions Successfully")
	return nil
}

func (s Service) createActions(ctx context.Context, actions []action.Action) error {
	for _, action := range actions {
		if _, err := s.actionService.Create(ctx, action); err != nil {
			return err
		}
	}
	return nil
}

func (s Service) bootstrapRoles(ctx context.Context) error {
	rolesList := []role.Role{
		role.DefinitionOrganizationAdmin,
		role.DefinitionProjectAdmin,
		role.DefinitionTeamAdmin,
		role.DefinitionTeamMember,
		role.DefinitionProjectAdmin,
	}
	if err := s.createRoles(ctx, rolesList); err != nil {
		return err
	}
	s.logger.Info("Bootstrap Roles Successfully")
	return nil
}

func (s Service) createRoles(ctx context.Context, roles []role.Role) error {
	for _, role := range roles {
		if _, err := s.roleService.Create(ctx, role); err != nil {
			return err
		}
	}
	return nil
}

func (s Service) createNamespaces(ctx context.Context, namespaces []namespace.Namespace) error {
	for _, ns := range namespaces {
		if _, err := s.namespaceService.Create(ctx, ns); err != nil {
			return err
		}
	}
	return nil
}

func (s Service) bootstrapNamespaces(ctx context.Context) error {
	namespaces := []namespace.Namespace{
		namespace.DefinitionOrg,
		namespace.DefinitionProject,
		namespace.DefinitionTeam,
		namespace.DefinitionUser,
	}

	if err := s.createNamespaces(ctx, namespaces); err != nil {
		return err
	}
	s.logger.Info("Bootstrap Namespaces Successfully")
	return nil
}
