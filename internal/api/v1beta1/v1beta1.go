package v1beta1

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/odpf/salt/server"
	"github.com/odpf/shield/internal/api"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
)

type Handler struct {
	shieldv1beta1.UnimplementedShieldServiceServer
	orgService          OrganizationService
	projectService      ProjectService
	groupService        GroupService
	roleService         RoleService
	policyService       PolicyService
	userService         UserService
	namespaceService    NamespaceService
	actionService       ActionService
	relationService     RelationService
	resourceService     ResourceService
	ruleService         RuleService
	identityProxyHeader string
}

var (
	internalServerError   = errors.New("internal server error")
	badRequestError       = errors.New("invalid syntax in body")
	permissionDeniedError = errors.New("permission denied")
)

func Register(ctx context.Context, s *server.MuxServer, gw *server.GRPCGateway, deps api.Deps) {
	gw.RegisterHandler(ctx, shieldv1beta1.RegisterShieldServiceHandlerFromEndpoint)

	s.RegisterService(
		&shieldv1beta1.ShieldService_ServiceDesc,
		&Handler{
			orgService:          deps.OrgService,
			projectService:      deps.ProjectService,
			groupService:        deps.GroupService,
			roleService:         deps.RoleService,
			policyService:       deps.PolicyService,
			userService:         deps.UserService,
			namespaceService:    deps.NamespaceService,
			actionService:       deps.ActionService,
			relationService:     deps.RelationService,
			resourceService:     deps.ResourceService,
			ruleService:         deps.RuleService,
			identityProxyHeader: deps.IdentityProxyHeader,
		},
	)
}

func isUUID(key string) bool {
	_, err := uuid.Parse(key)
	return err == nil
}
