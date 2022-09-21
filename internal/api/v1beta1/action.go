package v1beta1

import (
	"context"
	"errors"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

//go:generate mockery --name=ActionService -r --case underscore --with-expecter --structname ActionService --filename action_service.go --output=./mocks
type ActionService interface {
	Get(ctx context.Context, id string) (action.Action, error)
	List(ctx context.Context) ([]action.Action, error)
	Create(ctx context.Context, action action.Action) (action.Action, error)
	Update(ctx context.Context, id string, action action.Action) (action.Action, error)
}

var grpcActionNotFoundErr = status.Errorf(codes.NotFound, "action doesn't exist")

func (h Handler) ListActions(ctx context.Context, request *shieldv1beta1.ListActionsRequest) (*shieldv1beta1.ListActionsResponse, error) {
	logger := grpczap.Extract(ctx)

	actionsList, err := h.actionService.List(ctx)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	var actions []*shieldv1beta1.Action
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

func (h Handler) CreateAction(ctx context.Context, request *shieldv1beta1.CreateActionRequest) (*shieldv1beta1.CreateActionResponse, error) {
	logger := grpczap.Extract(ctx)

	newAction, err := h.actionService.Create(ctx, action.Action{
		ID:          request.GetBody().GetId(),
		Name:        request.GetBody().GetName(),
		NamespaceID: request.GetBody().GetNamespaceId(),
	})
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, namespace.ErrNotExist),
			errors.Is(err, action.ErrInvalidDetail),
			errors.Is(err, action.ErrInvalidID):
			return nil, grpcBadBodyError
		default:
			return nil, grpcInternalServerError
		}
	}

	actionPB, err := transformActionToPB(newAction)

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.CreateActionResponse{Action: &actionPB}, nil
}

func (h Handler) GetAction(ctx context.Context, request *shieldv1beta1.GetActionRequest) (*shieldv1beta1.GetActionResponse, error) {
	logger := grpczap.Extract(ctx)

	fetchedAction, err := h.actionService.Get(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, action.ErrNotExist), errors.Is(err, action.ErrInvalidID):
			return nil, grpcActionNotFoundErr
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

func (h Handler) UpdateAction(ctx context.Context, request *shieldv1beta1.UpdateActionRequest) (*shieldv1beta1.UpdateActionResponse, error) {
	logger := grpczap.Extract(ctx)

	updatedAction, err := h.actionService.Update(ctx, request.GetId(), action.Action{
		ID:          request.GetId(),
		Name:        request.GetBody().GetName(),
		NamespaceID: request.GetBody().GetNamespaceId(),
	})
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, action.ErrNotExist),
			errors.Is(err, action.ErrInvalidID):
			return nil, grpcActionNotFoundErr
		case errors.Is(err, namespace.ErrNotExist),
			errors.Is(err, action.ErrInvalidDetail):
			return nil, grpcBadBodyError
		default:
			return nil, grpcInternalServerError
		}
	}

	actionPB, err := transformActionToPB(updatedAction)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.UpdateActionResponse{Action: &actionPB}, nil
}

func transformActionToPB(act action.Action) (shieldv1beta1.Action, error) {
	return shieldv1beta1.Action{
		Id:        act.ID,
		Name:      act.Name,
		CreatedAt: timestamppb.New(act.CreatedAt),
		UpdatedAt: timestamppb.New(act.UpdatedAt),
	}, nil
}
