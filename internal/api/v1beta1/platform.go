package v1beta1

import (
	"context"

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
