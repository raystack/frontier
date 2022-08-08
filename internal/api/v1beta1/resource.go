package v1beta1

import (
	"context"
	"errors"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/core/resource"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ResourceService interface {
	Get(ctx context.Context, id string) (resource.Resource, error)
	List(ctx context.Context, flt resource.Filter) ([]resource.Resource, error)
	Create(ctx context.Context, resource resource.Resource) (resource.Resource, error)
	Update(ctx context.Context, id string, resource resource.Resource) (resource.Resource, error)
	CheckAuthz(ctx context.Context, resource resource.Resource, action action.Action) (bool, error)
}

var grpcResourceNotFoundErr = status.Errorf(codes.NotFound, "resource doesn't exist")

func (h Handler) ListResources(ctx context.Context, request *shieldv1beta1.ListResourcesRequest) (*shieldv1beta1.ListResourcesResponse, error) {
	logger := grpczap.Extract(ctx)
	var resources []*shieldv1beta1.Resource

	filters := resource.Filter{
		NamespaceID:    request.GetNamespaceId(),
		OrganizationID: request.GetOrganizationId(),
		ProjectID:      request.GetProjectId(),
		GroupID:        request.GetGroupId(),
	}

	resourcesList, err := h.resourceService.List(ctx, filters)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	for _, r := range resourcesList {
		resourcePB, err := transformResourceToPB(r)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}

		resources = append(resources, &resourcePB)
	}

	return &shieldv1beta1.ListResourcesResponse{
		Resources: resources,
	}, nil
}

func (h Handler) CreateResource(ctx context.Context, request *shieldv1beta1.CreateResourceRequest) (*shieldv1beta1.CreateResourceResponse, error) {
	logger := grpczap.Extract(ctx)
	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	newResource, err := h.resourceService.Create(ctx, resource.Resource{
		OrganizationID: request.GetBody().GetOrganizationId(),
		ProjectID:      request.GetBody().GetProjectId(),
		GroupID:        request.GetBody().GetGroupId(),
		NamespaceID:    request.GetBody().GetNamespaceId(),
		Name:           request.GetBody().GetName(),
		UserID:         request.GetBody().GetUserId(),
	})

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	resourcePB, err := transformResourceToPB(newResource)

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.CreateResourceResponse{
		Resource: &resourcePB,
	}, nil
}

func (h Handler) GetResource(ctx context.Context, request *shieldv1beta1.GetResourceRequest) (*shieldv1beta1.GetResourceResponse, error) {
	logger := grpczap.Extract(ctx)

	fetchedResource, err := h.resourceService.Get(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, resource.ErrNotExist):
			return nil, grpcResourceNotFoundErr
		case errors.Is(err, relation.ErrInvalidUUID):
			return nil, grpcBadBodyError
		default:
			return nil, grpcInternalServerError
		}
	}

	resourcePB, err := transformResourceToPB(fetchedResource)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.GetResourceResponse{
		Resource: &resourcePB,
	}, nil
}

func (h Handler) UpdateResource(ctx context.Context, request *shieldv1beta1.UpdateResourceRequest) (*shieldv1beta1.UpdateResourceResponse, error) {
	logger := grpczap.Extract(ctx)

	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	updatedResource, err := h.resourceService.Update(ctx, request.GetId(), resource.Resource{
		OrganizationID: request.GetBody().GetOrganizationId(),
		ProjectID:      request.GetBody().GetProjectId(),
		GroupID:        request.GetBody().GetGroupId(),
		NamespaceID:    request.GetBody().GetNamespaceId(),
		Name:           request.GetBody().GetName(),
		UserID:         request.GetBody().GetUserId(),
	})

	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, resource.ErrNotExist):
			return nil, grpcResourceNotFoundErr
		case errors.Is(err, relation.ErrInvalidUUID):
			return nil, grpcBadBodyError
		default:
			return nil, grpcInternalServerError
		}
	}

	resourcePB, err := transformResourceToPB(updatedResource)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.UpdateResourceResponse{
		Resource: &resourcePB,
	}, nil
}

func transformResourceToPB(from resource.Resource) (shieldv1beta1.Resource, error) {
	namespace, err := transformNamespaceToPB(from.Namespace)
	if err != nil {
		return shieldv1beta1.Resource{}, err
	}

	org, err := transformOrgToPB(from.Organization)
	if err != nil {
		return shieldv1beta1.Resource{}, err
	}

	project, err := transformProjectToPB(from.Project)
	if err != nil {
		return shieldv1beta1.Resource{}, err
	}

	group, err := transformGroupToPB(from.Group)
	if err != nil {
		return shieldv1beta1.Resource{}, err
	}

	user, err := transformUserToPB(from.User)
	if err != nil {
		return shieldv1beta1.Resource{}, err
	}

	return shieldv1beta1.Resource{
		Id:           from.Idxa,
		Urn:          from.URN,
		Name:         from.Name,
		Namespace:    &namespace,
		Organization: &org,
		Project:      &project,
		Group:        &group,
		User:         &user,
		CreatedAt:    timestamppb.New(from.CreatedAt),
		UpdatedAt:    timestamppb.New(from.UpdatedAt),
	}, nil
}
