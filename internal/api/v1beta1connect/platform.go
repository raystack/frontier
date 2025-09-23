package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"go.uber.org/zap"
)

func (h *ConnectHandler) AddPlatformUser(ctx context.Context, req *connect.Request[frontierv1beta1.AddPlatformUserRequest]) (*connect.Response[frontierv1beta1.AddPlatformUserResponse], error) {
	relationName := req.Msg.GetRelation()
	if !schema.IsPlatformRelation(relationName) {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	if req.Msg.GetUserId() != "" {
		if err := h.userService.Sudo(ctx, req.Msg.GetUserId(), relationName); err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	} else if req.Msg.GetServiceuserId() != "" {
		if err := h.serviceUserService.Sudo(ctx, req.Msg.GetServiceuserId(), relationName); err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	} else {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}
	return connect.NewResponse(&frontierv1beta1.AddPlatformUserResponse{}), nil
}

func (h *ConnectHandler) RemovePlatformUser(ctx context.Context, req *connect.Request[frontierv1beta1.RemovePlatformUserRequest]) (*connect.Response[frontierv1beta1.RemovePlatformUserResponse], error) {
	if req.Msg.GetUserId() != "" {
		if err := h.userService.UnSudo(ctx, req.Msg.GetUserId()); err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	} else if req.Msg.GetServiceuserId() != "" {
		if err := h.serviceUserService.UnSudo(ctx, req.Msg.GetServiceuserId()); err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	} else {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}
	return connect.NewResponse(&frontierv1beta1.RemovePlatformUserResponse{}), nil
}

func (h *ConnectHandler) ListPlatformUsers(ctx context.Context, req *connect.Request[frontierv1beta1.ListPlatformUsersRequest]) (*connect.Response[frontierv1beta1.ListPlatformUsersResponse], error) {
	logger := grpczap.Extract(ctx)
	relations, err := h.relationService.List(ctx, relation.Filter{
		Object: relation.Object{
			ID:        schema.PlatformID,
			Namespace: schema.PlatformNamespace,
		},
	})
	if err != nil {
		logger.Error("failed to list relations", zap.Error(err))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	subjectRelationMap := make(map[string]string)

	// fetch users
	userIDs := utils.Map(utils.Filter(relations, func(r relation.Relation) bool {
		return r.Subject.Namespace == schema.UserPrincipal
	}), func(r relation.Relation) string {
		subjectRelationMap[r.Subject.ID] = r.RelationName
		return r.Subject.ID
	})
	userPBs := make([]*frontierv1beta1.User, 0, len(userIDs))
	if len(userIDs) > 0 {
		users, err := h.userService.GetByIDs(ctx, userIDs)
		if err != nil {
			logger.Error("failed to get users by IDs", zap.Error(err))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		for _, u := range users {
			if u.Metadata == nil {
				u.Metadata = make(map[string]any)
			}
			u.Metadata["relation"] = subjectRelationMap[u.ID]
			userPB, err := transformUserToPB(u)
			if err != nil {
				logger.Error("failed to transform user to PB", zap.Error(err))
				return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
			}
			userPBs = append(userPBs, userPB)
		}
	}

	// fetch service users
	serviceUserIDs := utils.Map(utils.Filter(relations, func(r relation.Relation) bool {
		return r.Subject.Namespace == schema.ServiceUserPrincipal
	}), func(r relation.Relation) string {
		subjectRelationMap[r.Subject.ID] = r.RelationName
		return r.Subject.ID
	})
	serviceUserPBs := make([]*frontierv1beta1.ServiceUser, 0, len(serviceUserIDs))
	if len(serviceUserIDs) > 0 {
		serviceUsers, err := h.serviceUserService.GetByIDs(ctx, serviceUserIDs)
		if err != nil {
			logger.Error("failed to get service users by IDs", zap.Error(err))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		for _, u := range serviceUsers {
			if u.Metadata == nil {
				u.Metadata = make(map[string]any)
			}
			u.Metadata["relation"] = subjectRelationMap[u.ID]
			serviceUserPB, err := transformServiceUserToPB(u)
			if err != nil {
				logger.Error("failed to transform service user to PB", zap.Error(err))
				return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
			}
			serviceUserPBs = append(serviceUserPBs, serviceUserPB)
		}
	}

	return connect.NewResponse(&frontierv1beta1.ListPlatformUsersResponse{
		Users:        userPBs,
		Serviceusers: serviceUserPBs,
	}), nil
}
