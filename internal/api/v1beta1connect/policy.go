package v1beta1connect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/namespace"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PolicyService interface {
	Get(ctx context.Context, id string) (policy.Policy, error)
	List(ctx context.Context, f policy.Filter) ([]policy.Policy, error)
	Create(ctx context.Context, pol policy.Policy) (policy.Policy, error)
	Delete(ctx context.Context, id string) error
}

func (h *ConnectHandler) CreatePolicy(ctx context.Context, request *connect.Request[frontierv1beta1.CreatePolicyRequest]) (*connect.Response[frontierv1beta1.CreatePolicyResponse], error) {
	var metaDataMap metadata.Metadata
	if request.Msg.GetBody().GetMetadata() != nil {
		metaDataMap = metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())
	}

	resourceType, resourceID, err := schema.SplitNamespaceAndResourceID(request.Msg.GetBody().GetResource())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrNamespaceSplitNotation)
	}
	principalType, principalID, err := schema.SplitNamespaceAndResourceID(request.Msg.GetBody().GetPrincipal())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrNamespaceSplitNotation)
	}

	newPolicy, err := h.policyService.Create(ctx, policy.Policy{
		RoleID:        request.Msg.GetBody().GetRoleId(),
		ResourceID:    resourceID,
		ResourceType:  resourceType,
		PrincipalID:   principalID,
		PrincipalType: principalType,
		Metadata:      metaDataMap,
	})
	if err != nil {
		switch {
		case errors.Is(err, role.ErrInvalidID):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrInvalidRoleID)
		case errors.Is(err, policy.ErrInvalidDetail):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	policyPB, err := transformPolicyToPB(newPolicy)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	auditPolicyCreationEvent(ctx, newPolicy)
	return connect.NewResponse(&frontierv1beta1.CreatePolicyResponse{Policy: policyPB}), nil
}

func (h *ConnectHandler) GetPolicy(ctx context.Context, request *connect.Request[frontierv1beta1.GetPolicyRequest]) (*connect.Response[frontierv1beta1.GetPolicyResponse], error) {
	fetchedPolicy, err := h.policyService.Get(ctx, request.Msg.GetId())
	if err != nil {
		switch {
		case errors.Is(err, policy.ErrNotExist),
			errors.Is(err, policy.ErrInvalidUUID),
			errors.Is(err, policy.ErrInvalidID):
			return nil, connect.NewError(connect.CodeNotFound, ErrPolicyNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	policyPB, err := transformPolicyToPB(fetchedPolicy)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.GetPolicyResponse{Policy: policyPB}), nil
}

func (h *ConnectHandler) UpdatePolicy(ctx context.Context, request *connect.Request[frontierv1beta1.UpdatePolicyRequest]) (*connect.Response[frontierv1beta1.UpdatePolicyResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("unsupported at the moment"))
}

func (h *ConnectHandler) DeletePolicy(ctx context.Context, request *connect.Request[frontierv1beta1.DeletePolicyRequest]) (*connect.Response[frontierv1beta1.DeletePolicyResponse], error) {
	err := h.policyService.Delete(ctx, request.Msg.GetId())
	if err != nil {
		switch {
		case errors.Is(err, policy.ErrNotExist),
			errors.Is(err, policy.ErrInvalidID),
			errors.Is(err, policy.ErrInvalidUUID):
			return nil, connect.NewError(connect.CodeNotFound, ErrPolicyNotFound)
		case errors.Is(err, policy.ErrInvalidDetail),
			errors.Is(err, namespace.ErrNotExist):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		case errors.Is(err, policy.ErrConflict):
			return nil, connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	audit.GetAuditor(ctx, schema.PlatformOrgID.String()).Log(audit.PolicyDeletedEvent, audit.Target{
		ID:   request.Msg.GetId(),
		Type: "app/policy",
	})
	return connect.NewResponse(&frontierv1beta1.DeletePolicyResponse{}), nil
}

func (h *ConnectHandler) ListPolicies(ctx context.Context, request *connect.Request[frontierv1beta1.ListPoliciesRequest]) (*connect.Response[frontierv1beta1.ListPoliciesResponse], error) {
	var policies []*frontierv1beta1.Policy

	filter, err := h.resolveFilter(ctx, request.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	policyList, err := h.policyService.List(ctx, filter)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	for _, p := range policyList {
		policyPB, err := transformPolicyToPB(p)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		policies = append(policies, policyPB)
	}

	return connect.NewResponse(&frontierv1beta1.ListPoliciesResponse{Policies: policies}), nil
}

func transformPolicyToPB(policy policy.Policy) (*frontierv1beta1.Policy, error) {
	var metadata *structpb.Struct
	var err error
	if len(policy.Metadata) > 0 {
		metadata, err = structpb.NewStruct(policy.Metadata)
		if err != nil {
			return nil, err
		}
	}

	pbPol := &frontierv1beta1.Policy{
		Id:        policy.ID,
		RoleId:    policy.RoleID,
		Resource:  schema.JoinNamespaceAndResourceID(policy.ResourceType, policy.ResourceID),
		Principal: schema.JoinNamespaceAndResourceID(policy.PrincipalType, policy.PrincipalID),
		Metadata:  metadata,
	}
	if !policy.CreatedAt.IsZero() {
		pbPol.CreatedAt = timestamppb.New(policy.CreatedAt)
	}
	if !policy.UpdatedAt.IsZero() {
		pbPol.UpdatedAt = timestamppb.New(policy.UpdatedAt)
	}
	return pbPol, nil
}

func auditPolicyCreationEvent(ctx context.Context, policyCreated policy.Policy) {
	audit.GetAuditor(ctx, schema.PlatformOrgID.String()).
		LogWithAttrs(audit.PolicyCreatedEvent, audit.Target{
			ID:   policyCreated.ResourceID,
			Type: policyCreated.ResourceType,
		}, map[string]string{
			"role_id":        policyCreated.RoleID,
			"principal_id":   policyCreated.PrincipalID,
			"principal_type": policyCreated.PrincipalType,
		})
}

// resolveFilter resolves the filter from the request and returns a policy filter
// if the filter fields are not valid UUIDs, it will try to resolve them to their names and then return the filter. Note the group id is not resolved to name as it is not unique
func (h *ConnectHandler) resolveFilter(ctx context.Context, request *frontierv1beta1.ListPoliciesRequest) (policy.Filter, error) {
	var filter policy.Filter
	orgID := request.GetOrgId()
	if orgID != "" && !utils.IsValidUUID(orgID) {
		org, err := h.orgService.Get(ctx, orgID)
		if err != nil {
			return filter, err
		}
		orgID = org.ID
	}
	roleId := request.GetRoleId()
	if roleId != "" && !utils.IsValidUUID(roleId) {
		role, err := h.roleService.Get(ctx, roleId)
		if err != nil {
			return filter, err
		}
		roleId = role.ID
	}
	projectId := request.GetProjectId()
	if projectId != "" && !utils.IsValidUUID(projectId) {
		project, err := h.projectService.Get(ctx, projectId)
		if err != nil {
			return filter, err
		}
		projectId = project.ID
	}
	userId := request.GetUserId()
	if userId != "" && !utils.IsValidUUID(userId) {
		user, err := h.userService.GetByID(ctx, userId)
		if err != nil {
			return filter, err
		}
		userId = user.ID
	}
	return policy.Filter{
		PrincipalID: userId,
		OrgID:       orgID,
		ProjectID:   projectId,
		GroupID:     request.GetGroupId(),
		RoleID:      roleId,
	}, nil
}
