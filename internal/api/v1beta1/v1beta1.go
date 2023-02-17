package v1beta1

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/odpf/shield/internal/api"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	"google.golang.org/grpc"
)

type Handler struct {
	shieldv1beta1.UnimplementedShieldServiceServer
	orgService       OrganizationService
	projectService   ProjectService
	groupService     GroupService
	roleService      RoleService
	policyService    PolicyService
	userService      UserService
	namespaceService NamespaceService
	actionService    ActionService
	relationService  RelationService
	resourceService  ResourceService
	ruleService      RuleService
}

func Register(ctx context.Context, address string, s *grpc.Server, gw *runtime.ServeMux, deps api.Deps) {
	shieldv1beta1.RegisterShieldServiceHandlerFromEndpoint(ctx, gw, address, []grpc.DialOption{grpc.WithInsecure()})

	s.RegisterService(
		&shieldv1beta1.ShieldService_ServiceDesc,
		&Handler{
			orgService:       deps.OrgService,
			projectService:   deps.ProjectService,
			groupService:     deps.GroupService,
			roleService:      deps.RoleService,
			policyService:    deps.PolicyService,
			userService:      deps.UserService,
			namespaceService: deps.NamespaceService,
			actionService:    deps.ActionService,
			relationService:  deps.RelationService,
			resourceService:  deps.ResourceService,
			ruleService:      deps.RuleService,
		},
	)
}
