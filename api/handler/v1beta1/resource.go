package v1beta1

import (
	"context"
	"errors"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/odpf/shield/internal/relation"
	"github.com/odpf/shield/internal/resource"
	"github.com/odpf/shield/model"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ResourceService interface {
	Get(ctx context.Context, id string) (model.Resource, error)
	List(ctx context.Context, filters model.ResourceFilters) ([]model.Resource, error)
	Create(ctx context.Context, resource model.Resource) (model.Resource, error)
	Update(ctx context.Context, id string, resource model.Resource) (model.Resource, error)
}

var grpcResourceNotFoundErr = status.Errorf(codes.NotFound, "resource doesn't exist")

func (v Dep) ListResources(ctx context.Context, request *shieldv1beta1.ListResourcesRequest) (*shieldv1beta1.ListResourcesResponse, error) {
	logger := grpczap.Extract(ctx)
	var resources []*shieldv1beta1.Resource

	filters := model.ResourceFilters{
		NamespaceId:    request.NamespaceId,
		OrganizationId: request.OrganizationId,
		ProjectId:      request.ProjectId,
		GroupId:        request.GroupId,
	}

	resourcesList, err := v.ResourceService.List(ctx, filters)
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

func (v Dep) CreateResource(ctx context.Context, request *shieldv1beta1.CreateResourceRequest) (*shieldv1beta1.CreateResourceResponse, error) {
	logger := grpczap.Extract(ctx)
	if request.Body == nil {
		return nil, grpcBadBodyError
	}

	newResource, err := v.ResourceService.Create(ctx, model.Resource{
		OrganizationId: request.GetBody().OrganizationId,
		ProjectId:      request.GetBody().ProjectId,
		GroupId:        request.GetBody().GroupId,
		NamespaceId:    request.GetBody().NamespaceId,
		Name:           request.GetBody().Name,
		UserId:         request.GetBody().UserId,
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

func (v Dep) GetResource(ctx context.Context, request *shieldv1beta1.GetResourceRequest) (*shieldv1beta1.GetResourceResponse, error) {
	logger := grpczap.Extract(ctx)

	fetchedResource, err := v.ResourceService.Get(ctx, request.Id)

	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, resource.ResourceDoesntExist):
			return nil, grpcResourceNotFoundErr
		case errors.Is(err, relation.InvalidUUID):
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

func (v Dep) UpdateResource(ctx context.Context, request *shieldv1beta1.UpdateResourceRequest) (*shieldv1beta1.UpdateResourceResponse, error) {
	logger := grpczap.Extract(ctx)

	if request.Body == nil {
		return nil, grpcBadBodyError
	}

	updatedResource, err := v.ResourceService.Update(ctx, request.Id, model.Resource{
		OrganizationId: request.GetBody().OrganizationId,
		ProjectId:      request.GetBody().ProjectId,
		GroupId:        request.GetBody().GroupId,
		NamespaceId:    request.GetBody().NamespaceId,
		Name:           request.GetBody().Name,
		UserId:         request.GetBody().UserId,
	})

	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, resource.ResourceDoesntExist):
			return nil, grpcResourceNotFoundErr
		case errors.Is(err, relation.InvalidUUID):
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

func transformResourceToPB(from model.Resource) (shieldv1beta1.Resource, error) {
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
		Urn:          from.Urn,
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
