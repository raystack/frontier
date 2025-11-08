package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"go.uber.org/zap"
)

func (h *ConnectHandler) AddPlatformUser(ctx context.Context, req *connect.Request[frontierv1beta1.AddPlatformUserRequest]) (*connect.Response[frontierv1beta1.AddPlatformUserResponse], error) {
	errorLogger := NewErrorLogger()
	relationName := req.Msg.GetRelation()

	if !schema.IsPlatformRelation(relationName) {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	if req.Msg.GetUserId() != "" {
		if err := h.userService.Sudo(ctx, req.Msg.GetUserId(), relationName); err != nil {
			errorLogger.LogServiceError(ctx, req, "AddPlatformUser.UserSudo", err,
				zap.String("user_id", req.Msg.GetUserId()),
				zap.String("relation", relationName))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	} else if req.Msg.GetServiceuserId() != "" {
		if err := h.serviceUserService.Sudo(ctx, req.Msg.GetServiceuserId(), relationName); err != nil {
			errorLogger.LogServiceError(ctx, req, "AddPlatformUser.ServiceUserSudo", err,
				zap.String("service_user_id", req.Msg.GetServiceuserId()),
				zap.String("relation", relationName))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	} else {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}
	return connect.NewResponse(&frontierv1beta1.AddPlatformUserResponse{}), nil
}

func (h *ConnectHandler) RemovePlatformUser(ctx context.Context, req *connect.Request[frontierv1beta1.RemovePlatformUserRequest]) (*connect.Response[frontierv1beta1.RemovePlatformUserResponse], error) {
	errorLogger := NewErrorLogger()

	if req.Msg.GetUserId() != "" {
		if err := h.userService.UnSudo(ctx, req.Msg.GetUserId()); err != nil {
			errorLogger.LogServiceError(ctx, req, "RemovePlatformUser.UserUnSudo", err,
				zap.String("user_id", req.Msg.GetUserId()))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	} else if req.Msg.GetServiceuserId() != "" {
		if err := h.serviceUserService.UnSudo(ctx, req.Msg.GetServiceuserId()); err != nil {
			errorLogger.LogServiceError(ctx, req, "RemovePlatformUser.ServiceUserUnSudo", err,
				zap.String("service_user_id", req.Msg.GetServiceuserId()))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	} else {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}
	return connect.NewResponse(&frontierv1beta1.RemovePlatformUserResponse{}), nil
}

func (h *ConnectHandler) ListPlatformUsers(ctx context.Context, req *connect.Request[frontierv1beta1.ListPlatformUsersRequest]) (*connect.Response[frontierv1beta1.ListPlatformUsersResponse], error) {
	errorLogger := NewErrorLogger()
	relations, err := h.relationService.List(ctx, relation.Filter{
		Object: relation.Object{
			ID:        schema.PlatformID,
			Namespace: schema.PlatformNamespace,
		},
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, req, "ListPlatformUsers.ListRelations", err)
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
			errorLogger.LogServiceError(ctx, req, "ListPlatformUsers.GetUsersByIDs", err,
				zap.Strings("user_ids", userIDs))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		for _, u := range users {
			if u.Metadata == nil {
				u.Metadata = make(map[string]any)
			}
			u.Metadata["relation"] = subjectRelationMap[u.ID]
			userPB, err := transformUserToPB(u)
			if err != nil {
				errorLogger.LogTransformError(ctx, req, "ListPlatformUsers.TransformUser", u.ID, err)
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
			errorLogger.LogServiceError(ctx, req, "ListPlatformUsers.GetServiceUsersByIDs", err,
				zap.Strings("service_user_ids", serviceUserIDs))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		for _, u := range serviceUsers {
			if u.Metadata == nil {
				u.Metadata = make(map[string]any)
			}
			u.Metadata["relation"] = subjectRelationMap[u.ID]
			serviceUserPB, err := transformServiceUserToPB(u)
			if err != nil {
				errorLogger.LogTransformError(ctx, req, "ListPlatformUsers.TransformServiceUser", u.ID, err)
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
