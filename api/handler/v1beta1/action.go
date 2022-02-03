package v1beta1

import (
	"context"
	"errors"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/odpf/shield/internal/schema"
	"github.com/odpf/shield/model"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ActionService interface {
	GetAction(ctx context.Context, id string) (model.Action, error)
	ListActions(ctx context.Context) ([]model.Action, error)
	CreateAction(ctx context.Context, action model.Action) (model.Action, error)
	UpdateAction(ctx context.Context, id string, action model.Action) (model.Action, error)
}

var grpcActionNotFoundErr = status.Errorf(codes.NotFound, "action doesn't exist")

func (v Dep) ListActions(ctx context.Context, request *shieldv1beta1.ListActionsRequest) (*shieldv1beta1.ListActionsResponse, error) {
	logger := grpczap.Extract(ctx)
	var actions []*shieldv1beta1.Action

	actionsList, err := v.ActionService.ListActions(ctx)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	for _, act := range actionsList {
		actPB, err := transformActionToPB(act)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}

		actions = append(actions, &actPB)
	}

	return &shieldv1beta1.ListActionsResponse{Actions: actions}, nil
}

func (v Dep) CreateAction(ctx context.Context, request *shieldv1beta1.CreateActionRequest) (*shieldv1beta1.CreateActionResponse, error) {
	logger := grpczap.Extract(ctx)

	newAction, err := v.ActionService.CreateAction(ctx, model.Action{
		Id:          request.GetBody().Id,
		Name:        request.GetBody().Name,
		NamespaceId: request.GetBody().NamespaceId,
	})

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	actionPB, err := transformActionToPB(newAction)

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.CreateActionResponse{Action: &actionPB}, nil
}

func (v Dep) GetAction(ctx context.Context, request *shieldv1beta1.GetActionRequest) (*shieldv1beta1.GetActionResponse, error) {
	logger := grpczap.Extract(ctx)

	fetchedAction, err := v.ActionService.GetAction(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, schema.ActionDoesntExist):
			return nil, grpcActionNotFoundErr
		case errors.Is(err, schema.InvalidUUID):
			return nil, grpcBadBodyError
		default:
			return nil, grpcInternalServerError
		}
	}

	actionPB, err := transformActionToPB(fetchedAction)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.GetActionResponse{Action: &actionPB}, nil
}

func (v Dep) UpdateAction(ctx context.Context, request *shieldv1beta1.UpdateActionRequest) (*shieldv1beta1.UpdateActionResponse, error) {
	logger := grpczap.Extract(ctx)

	updatedAction, err := v.ActionService.UpdateAction(ctx, request.GetId(), model.Action{
		Id:          request.GetId(),
		Name:        request.GetBody().Name,
		NamespaceId: request.GetBody().NamespaceId,
	})

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	actionPB, err := transformActionToPB(updatedAction)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.UpdateActionResponse{Action: &actionPB}, nil
}

func transformActionToPB(act model.Action) (shieldv1beta1.Action, error) {
	act.Namespace.Id = act.NamespaceId
	namespace, err := transformNamespaceToPB(act.Namespace)
	if err != nil {
		return shieldv1beta1.Action{}, err
	}
	return shieldv1beta1.Action{
		Id:        act.Id,
		Name:      act.Name,
		Namespace: &namespace,
		CreatedAt: timestamppb.New(act.CreatedAt),
		UpdatedAt: timestamppb.New(act.UpdatedAt),
	}, nil
}
