package v1beta1connect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/metadata"
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
