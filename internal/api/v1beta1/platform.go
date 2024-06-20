package v1beta1

import (
	"context"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/raystack/frontier/internal/bootstrap/schema"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

func (h Handler) AddPlatformUser(ctx context.Context, req *frontierv1beta1.AddPlatformUserRequest) (*frontierv1beta1.AddPlatformUserResponse, error) {
	relationName := req.GetRelation()
	if !schema.IsPlatformRelation(relationName) {
		return nil, grpcBadBodyError
	}

	if req.GetUserId() != "" {
		if err := h.userService.Sudo(ctx, req.GetUserId(), relationName); err != nil {
			return nil, err
		}
	} else if req.GetServiceuserId() != "" {
		if err := h.serviceUserService.Sudo(ctx, req.GetServiceuserId(), relationName); err != nil {
			return nil, err
		}
	} else {
		return nil, grpcBadBodyError
	}
	return &frontierv1beta1.AddPlatformUserResponse{}, nil
}

func (h Handler) RemovePlatformUser(ctx context.Context, req *frontierv1beta1.RemovePlatformUserRequest) (*frontierv1beta1.RemovePlatformUserResponse, error) {
	if req.GetUserId() != "" {
		if err := h.userService.UnSudo(ctx, req.GetUserId()); err != nil {
			return nil, err
		}
	} else if req.GetServiceuserId() != "" {
		if err := h.serviceUserService.UnSudo(ctx, req.GetServiceuserId()); err != nil {
			return nil, err
		}
	} else {
		return nil, grpcBadBodyError
	}
	return &frontierv1beta1.RemovePlatformUserResponse{}, nil
}

func (h Handler) ListPlatformUsers(ctx context.Context, req *frontierv1beta1.ListPlatformUsersRequest) (*frontierv1beta1.ListPlatformUsersResponse, error) {
	logger := grpczap.Extract(ctx)
	relations, err := h.relationService.List(ctx, relation.Filter{
		Object: relation.Object{
			ID:        schema.PlatformID,
			Namespace: schema.PlatformNamespace,
		},
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	// fetch users
	userIDs := utils.Map(utils.Filter(relations, func(r relation.Relation) bool {
		return r.Subject.Namespace == schema.UserPrincipal
	}), func(r relation.Relation) string {
		return r.Subject.ID
	})
	userPBs := make([]*frontierv1beta1.User, 0, len(userIDs))
	if len(userIDs) > 0 {
		users, err := h.userService.GetByIDs(ctx, userIDs)
		if err != nil {
			logger.Error(err.Error())
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		for _, u := range users {
			userPB, err := transformUserToPB(u)
			if err != nil {
				logger.Error(err.Error())
				return nil, grpcInternalServerError
			}
			userPBs = append(userPBs, userPB)
		}
	}

	// fetch service users
	serviceUserIDs := utils.Map(utils.Filter(relations, func(r relation.Relation) bool {
		return r.Subject.Namespace == schema.ServiceUserPrincipal
	}), func(r relation.Relation) string {
		return r.Subject.ID
	})
	serviceUserPBs := make([]*frontierv1beta1.ServiceUser, 0, len(serviceUserIDs))
	if len(serviceUserIDs) > 0 {
		serviceUsers, err := h.serviceUserService.GetByIDs(ctx, serviceUserIDs)
		if err != nil {
			logger.Error(err.Error())
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		for _, u := range serviceUsers {
			serviceUserPB, err := transformServiceUserToPB(u)
			if err != nil {
				logger.Error(err.Error())
				return nil, grpcInternalServerError
			}
			serviceUserPBs = append(serviceUserPBs, serviceUserPB)
		}
	}

	return &frontierv1beta1.ListPlatformUsersResponse{
		Users:        userPBs,
		Serviceusers: serviceUserPBs,
	}, nil
}
