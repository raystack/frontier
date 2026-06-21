package v1beta1connect

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

func (h *ConnectHandler) AddPlatformUser(ctx context.Context, req *connect.Request[frontierv1beta1.AddPlatformUserRequest]) (*connect.Response[frontierv1beta1.AddPlatformUserResponse], error) {
	relationName := req.Msg.GetRelation()

	if !schema.IsPlatformRelation(relationName) {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	if req.Msg.GetUserId() != "" {
		if err := h.userService.Sudo(ctx, req.Msg.GetUserId(), relationName); err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("AddPlatformUser.UserSudo: user_id=%s relation=%s: %w", req.Msg.GetUserId(), relationName, err))
		}
	} else if req.Msg.GetServiceuserId() != "" {
		if err := h.serviceUserService.Sudo(ctx, req.Msg.GetServiceuserId(), relationName); err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("AddPlatformUser.ServiceUserSudo: service_user_id=%s relation=%s: %w", req.Msg.GetServiceuserId(), relationName, err))
		}
	} else {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}
	return connect.NewResponse(&frontierv1beta1.AddPlatformUserResponse{}), nil
}

func (h *ConnectHandler) RemovePlatformUser(ctx context.Context, req *connect.Request[frontierv1beta1.RemovePlatformUserRequest]) (*connect.Response[frontierv1beta1.RemovePlatformUserResponse], error) {
	// Remove the principal from the platform entirely: strip both the admin
	// (superuser) and member (check) relations. Each UnSudo is a no-op for a
	// relation the principal doesn't hold.
	platformRelations := []string{schema.AdminRelationName, schema.MemberRelationName}

	if req.Msg.GetUserId() != "" {
		for _, relationName := range platformRelations {
			if err := h.userService.UnSudo(ctx, req.Msg.GetUserId(), relationName); err != nil {
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("RemovePlatformUser.UserUnSudo: user_id=%s relation=%s: %w", req.Msg.GetUserId(), relationName, err))
			}
		}
	} else if req.Msg.GetServiceuserId() != "" {
		for _, relationName := range platformRelations {
			if err := h.serviceUserService.UnSudo(ctx, req.Msg.GetServiceuserId(), relationName); err != nil {
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("RemovePlatformUser.ServiceUserUnSudo: service_user_id=%s relation=%s: %w", req.Msg.GetServiceuserId(), relationName, err))
			}
		}
	} else {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}
	return connect.NewResponse(&frontierv1beta1.RemovePlatformUserResponse{}), nil
}

func (h *ConnectHandler) ListPlatformUsers(ctx context.Context, req *connect.Request[frontierv1beta1.ListPlatformUsersRequest]) (*connect.Response[frontierv1beta1.ListPlatformUsersResponse], error) {
	relations, err := h.relationService.List(ctx, relation.Filter{
		Object: relation.Object{
			ID:        schema.PlatformID,
			Namespace: schema.PlatformNamespace,
		},
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListPlatformUsers.ListRelations: %w", err))
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
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListPlatformUsers.GetUsersByIDs: user_ids=%v: %w", userIDs, err))
		}
		for _, u := range users {
			if u.Metadata == nil {
				u.Metadata = make(map[string]any)
			}
			u.Metadata["relation"] = subjectRelationMap[u.ID]
			userPB, err := transformUserToPB(u)
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListPlatformUsers.TransformUser: entity_id=%s: %w", u.ID, err))
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
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListPlatformUsers.GetServiceUsersByIDs: service_user_ids=%v: %w", serviceUserIDs, err))
		}
		for _, u := range serviceUsers {
			if u.Metadata == nil {
				u.Metadata = make(map[string]any)
			}
			u.Metadata["relation"] = subjectRelationMap[u.ID]
			serviceUserPB, err := transformServiceUserToPB(u)
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListPlatformUsers.TransformServiceUser: entity_id=%s: %w", u.ID, err))
			}
			serviceUserPBs = append(serviceUserPBs, serviceUserPB)
		}
	}

	return connect.NewResponse(&frontierv1beta1.ListPlatformUsersResponse{
		Users:        userPBs,
		Serviceusers: serviceUserPBs,
	}), nil
}
