package v1

import (
	"context"
	"errors"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/odpf/shield/internal/schema"
	"github.com/odpf/shield/model"
	shieldv1 "go.buf.build/odpf/gwv/odpf/proton/odpf/shield/v1"
)

type PolicyService interface {
	GetPolicy(ctx context.Context, id string) (model.Policy, error)
	ListPolicies(ctx context.Context) ([]model.Policy, error)
	CreatePolicy(ctx context.Context, policy model.Policy) ([]model.Policy, error)
	UpdatePolicy(ctx context.Context, id string, policy model.Policy) ([]model.Policy, error)
}

var grpcPolicyNotFoundErr = status.Errorf(codes.NotFound, "policy doesn't exist")

func (v Dep) ListPolicies(ctx context.Context, request *shieldv1.ListPoliciesRequest) (*shieldv1.ListPoliciesResponse, error) {
	logger := grpczap.Extract(ctx)
	var policies []*shieldv1.Policy

	policyList, err := v.PolicyService.ListPolicies(ctx)
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

	return &shieldv1.ListPoliciesResponse{Policies: policies}, nil
}

func (v Dep) CreatePolicy(ctx context.Context, request *shieldv1.CreatePolicyRequest) (*shieldv1.CreatePolicyResponse, error) {
	logger := grpczap.Extract(ctx)
	var policies []*shieldv1.Policy

	newPolicies, err := v.PolicyService.CreatePolicy(ctx, model.Policy{
		RoleId:      request.GetBody().RoleId,
		NamespaceId: request.GetBody().NamespaceId,
		ActionId:    request.GetBody().ActionId,
	})

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	for _, p := range newPolicies {
		policyPB, err := transformPolicyToPB(p)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}

		policies = append(policies, &policyPB)
	}

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1.CreatePolicyResponse{Policies: policies}, nil
}

func (v Dep) GetPolicy(ctx context.Context, request *shieldv1.GetPolicyRequest) (*shieldv1.GetPolicyResponse, error) {
	logger := grpczap.Extract(ctx)

	fetchedPolicy, err := v.PolicyService.GetPolicy(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, schema.PolicyDoesntExist):
			return nil, grpcPolicyNotFoundErr
		case errors.Is(err, schema.InvalidUUID):
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

	return &shieldv1.GetPolicyResponse{Policy: &policyPB}, nil
}

func (v Dep) UpdatePolicy(ctx context.Context, request *shieldv1.UpdatePolicyRequest) (*shieldv1.UpdatePolicyResponse, error) {
	logger := grpczap.Extract(ctx)
	var policies []*shieldv1.Policy

	updatedPolices, err := v.PolicyService.UpdatePolicy(ctx, request.GetId(), model.Policy{
		RoleId:      request.GetBody().RoleId,
		NamespaceId: request.GetBody().NamespaceId,
		ActionId:    request.GetBody().ActionId,
	})

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	for _, p := range updatedPolices {
		policyPB, err := transformPolicyToPB(p)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}

		policies = append(policies, &policyPB)
	}

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	return &shieldv1.UpdatePolicyResponse{Policies: policies}, nil
}

func transformPolicyToPB(policy model.Policy) (shieldv1.Policy, error) {
	role, err := transformRoleToPB(policy.Role)
	if err != nil {
		return shieldv1.Policy{}, err
	}

	action, err := transformActionToPB(policy.Action)

	if err != nil {
		return shieldv1.Policy{}, err
	}

	namespace, err := transformNamespaceToPB(policy.Namespace)

	if err != nil {
		return shieldv1.Policy{}, err
	}

	return shieldv1.Policy{
		Id:        policy.Id,
		Role:      &role,
		Action:    &action,
		Namespace: &namespace,
		CreatedAt: timestamppb.New(policy.CreatedAt),
		UpdatedAt: timestamppb.New(policy.UpdatedAt),
	}, nil
}
