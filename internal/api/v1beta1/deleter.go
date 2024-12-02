package v1beta1

import (
	"context"

	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CascadeDeleter interface {
	DeleteProject(ctx context.Context, id string) error
	DeleteOrganization(ctx context.Context, id string) error
	RemoveUsersFromOrg(ctx context.Context, orgID string, userIDs []string) error
	DeleteUser(ctx context.Context, userID string) error
}

func (h Handler) DeleteProject(ctx context.Context, request *frontierv1beta1.DeleteProjectRequest) (*frontierv1beta1.DeleteProjectResponse, error) {
	if err := h.deleterService.DeleteProject(ctx, request.GetId()); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &frontierv1beta1.DeleteProjectResponse{}, nil
}

func (h Handler) DeleteOrganization(ctx context.Context, request *frontierv1beta1.DeleteOrganizationRequest) (*frontierv1beta1.DeleteOrganizationResponse, error) {
	if err := h.deleterService.DeleteOrganization(ctx, request.GetId()); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &frontierv1beta1.DeleteOrganizationResponse{}, nil
}
