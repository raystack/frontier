package v1beta1

import (
	"context"
	"errors"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/odpf/shield/core/policy"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
)

type PolicyService interface {
	Get(ctx context.Context, id string) (policy.Policy, error)
	List(ctx context.Context) ([]policy.Policy, error)
	Create(ctx context.Context, pol policy.Policy) ([]policy.Policy, error)
	Update(ctx context.Context, pol policy.Policy) ([]policy.Policy, error)
}

var grpcPolicyNotFoundErr = status.Errorf(codes.NotFound, "policy doesn't exist")

func (h Handler) ListPolicies(ctx context.Context, request *shieldv1beta1.ListPoliciesRequest) (*shieldv1beta1.ListPoliciesResponse, error) {
	logger := grpczap.Extract(ctx)
	var policies []*shieldv1beta1.Policy

	policyList, err := h.policyService.List(ctx)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	for _, p := range policyList {
		policyPB, err := transformPolicyToPB(p)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}

		policies = append(policies, &policyPB)
	}

	return &shieldv1beta1.ListPoliciesResponse{Policies: policies}, nil
}

func (h Handler) CreatePolicy(ctx context.Context, request *shieldv1beta1.CreatePolicyRequest) (*shieldv1beta1.CreatePolicyResponse, error) {
	logger := grpczap.Extract(ctx)
	var policies []*shieldv1beta1.Policy

	newPolicies, err := h.policyService.Create(ctx, policy.Policy{
		RoleID:      request.GetBody().GetRoleId(),
		NamespaceID: request.GetBody().GetNamespaceId(),
		ActionID:    request.GetBody().GetActionId(),
	})

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	for _, p := range newPolicies {
		policyPB, err := transformPolicyToPB(p)
		if err != nil {
			logger.Error(err.Error())
			switch {
			case errors.Is(err, policy.ErrNotExist):
				return nil, grpcPolicyNotFoundErr
			default:
				return nil, grpcInternalServerError
			}
		}

		policies = append(policies, &policyPB)
	}

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.CreatePolicyResponse{Policies: policies}, nil
}

func (h Handler) GetPolicy(ctx context.Context, request *shieldv1beta1.GetPolicyRequest) (*shieldv1beta1.GetPolicyResponse, error) {
	logger := grpczap.Extract(ctx)

	fetchedPolicy, err := h.policyService.Get(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, policy.ErrNotExist):
			return nil, grpcPolicyNotFoundErr
		case errors.Is(err, policy.ErrInvalidUUID), errors.Is(err, policy.ErrInvalidID):
			return nil, grpcBadBodyError
		default:
			return nil, grpcInternalServerError
		}
	}

	policyPB, err := transformPolicyToPB(fetchedPolicy)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.GetPolicyResponse{Policy: &policyPB}, nil
}

func (h Handler) UpdatePolicy(ctx context.Context, request *shieldv1beta1.UpdatePolicyRequest) (*shieldv1beta1.UpdatePolicyResponse, error) {
	logger := grpczap.Extract(ctx)
	var policies []*shieldv1beta1.Policy

	updatedPolices, err := h.policyService.Update(ctx, policy.Policy{
		ID:          request.GetId(),
		RoleID:      request.GetBody().GetRoleId(),
		NamespaceID: request.GetBody().GetNamespaceId(),
		ActionID:    request.GetBody().GetActionId(),
	})

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	for _, p := range updatedPolices {
		policyPB, err := transformPolicyToPB(p)
		if err != nil {
			logger.Error(err.Error())
			switch {
			case errors.Is(err, policy.ErrNotExist):
				return nil, grpcPolicyNotFoundErr
			case errors.Is(err, policy.ErrConflict):
				return nil, grpcConflictError
			default:
				return nil, grpcInternalServerError
			}
		}

		policies = append(policies, &policyPB)
	}

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	return &shieldv1beta1.UpdatePolicyResponse{Policies: policies}, nil
}

func transformPolicyToPB(policy policy.Policy) (shieldv1beta1.Policy, error) {
	role, err := transformRoleToPB(policy.Role)
	if err != nil {
		return shieldv1beta1.Policy{}, err
	}

	action, err := transformActionToPB(policy.Action)

	if err != nil {
		return shieldv1beta1.Policy{}, err
	}

	namespace, err := transformNamespaceToPB(policy.Namespace)

	if err != nil {
		return shieldv1beta1.Policy{}, err
	}

	return shieldv1beta1.Policy{
		Id:        policy.ID,
		Role:      &role,
		Action:    &action,
		Namespace: &namespace,
		CreatedAt: timestamppb.New(policy.CreatedAt),
		UpdatedAt: timestamppb.New(policy.UpdatedAt),
	}, nil
}
