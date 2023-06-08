package v1beta1

import (
	"context"
	"errors"

	"github.com/odpf/shield/internal/bootstrap/schema"

	"github.com/odpf/shield/core/relation"

	"google.golang.org/protobuf/types/known/structpb"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/odpf/shield/core/resource"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/pkg/metadata"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

//go:generate mockery --name=ResourceService -r --case underscore --with-expecter --structname ResourceService --filename resource_service.go --output=./mocks
type ResourceService interface {
	Get(ctx context.Context, id string) (resource.Resource, error)
	List(ctx context.Context, flt resource.Filter) ([]resource.Resource, error)
	Create(ctx context.Context, resource resource.Resource) (resource.Resource, error)
	Update(ctx context.Context, resource resource.Resource) (resource.Resource, error)
	CheckAuthz(ctx context.Context, rel relation.Object, permissionName string) (bool, error)
}

var grpcResourceNotFoundErr = status.Errorf(codes.NotFound, "resource doesn't exist")

func (h Handler) ListResources(ctx context.Context, request *shieldv1beta1.ListResourcesRequest) (*shieldv1beta1.ListResourcesResponse, error) {
	logger := grpczap.Extract(ctx)
	var resources []*shieldv1beta1.Resource
	namespaceID := schema.ParseNamespaceAliasIfRequired(request.GetNamespace())
	filters := resource.Filter{
		NamespaceID: namespaceID,
		ProjectID:   request.GetProjectId(),
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
		resources = append(resources, resourcePB)
	}

	return &shieldv1beta1.ListResourcesResponse{
		Resources: resources,
	}, nil
}

func (h Handler) ListProjectResources(ctx context.Context, request *shieldv1beta1.ListProjectResourcesRequest) (*shieldv1beta1.ListProjectResourcesResponse, error) {
	logger := grpczap.Extract(ctx)

	var resources []*shieldv1beta1.Resource
	namespaceID := schema.ParseNamespaceAliasIfRequired(request.GetNamespace())
	filters := resource.Filter{
		NamespaceID: namespaceID,
		ProjectID:   request.GetProjectId(),
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
		resources = append(resources, resourcePB)
	}

	return &shieldv1beta1.ListProjectResourcesResponse{
		Resources: resources,
	}, nil
}

func (h Handler) CreateProjectResource(ctx context.Context, request *shieldv1beta1.CreateProjectResourceRequest) (*shieldv1beta1.CreateProjectResourceResponse, error) {
	logger := grpczap.Extract(ctx)
	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	var metaDataMap metadata.Metadata
	var err error
	if request.GetBody().GetMetadata() != nil {
		metaDataMap, err = metadata.Build(request.GetBody().GetMetadata().AsMap())
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcBadBodyError
		}
	}

	namespaceID := schema.ParseNamespaceAliasIfRequired(request.GetBody().GetNamespace())
	newResource, err := h.resourceService.Create(ctx, resource.Resource{
		ID:          request.GetId(),
		Name:        request.GetBody().GetName(),
		ProjectID:   request.GetProjectId(),
		NamespaceID: namespaceID,
		UserID:      request.GetBody().GetUserId(),
		Metadata:    metaDataMap,
	})
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, user.ErrInvalidEmail):
			return nil, grpcUnauthenticated
		case errors.Is(err, resource.ErrInvalidUUID),
			errors.Is(err, resource.ErrInvalidDetail):
			return nil, grpcBadBodyError
		default:
			return nil, grpcInternalServerError
		}
	}

	resourcePB, err := transformResourceToPB(newResource)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.CreateProjectResourceResponse{
		Resource: resourcePB,
	}, nil
}

func (h Handler) GetProjectResource(ctx context.Context, request *shieldv1beta1.GetProjectResourceRequest) (*shieldv1beta1.GetProjectResourceResponse, error) {
	logger := grpczap.Extract(ctx)

	fetchedResource, err := h.resourceService.Get(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, resource.ErrNotExist),
			errors.Is(err, resource.ErrInvalidUUID),
			errors.Is(err, resource.ErrInvalidID):
			return nil, grpcResourceNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}

	resourcePB, err := transformResourceToPB(fetchedResource)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.GetProjectResourceResponse{
		Resource: resourcePB,
	}, nil
}

func (h Handler) UpdateProjectResource(ctx context.Context, request *shieldv1beta1.UpdateProjectResourceRequest) (*shieldv1beta1.UpdateProjectResourceResponse, error) {
	logger := grpczap.Extract(ctx)
	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	namespaceID := schema.ParseNamespaceAliasIfRequired(request.GetBody().GetNamespace())
	updatedResource, err := h.resourceService.Update(ctx, resource.Resource{
		ID:          request.GetId(),
		ProjectID:   request.GetProjectId(),
		NamespaceID: namespaceID,
		Name:        request.GetBody().GetName(),
		UserID:      request.GetBody().GetUserId(),
	})
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, resource.ErrNotExist),
			errors.Is(err, resource.ErrInvalidUUID),
			errors.Is(err, resource.ErrInvalidID):
			return nil, grpcResourceNotFoundErr
		case errors.Is(err, resource.ErrInvalidDetail),
			errors.Is(err, resource.ErrInvalidURN):
			return nil, grpcBadBodyError
		case errors.Is(err, resource.ErrConflict):
			return nil, grpcConflictError
		default:
			return nil, grpcInternalServerError
		}
	}

	resourcePB, err := transformResourceToPB(updatedResource)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.UpdateProjectResourceResponse{
		Resource: resourcePB,
	}, nil
}

func transformResourceToPB(from resource.Resource) (*shieldv1beta1.Resource, error) {
	var metadata *structpb.Struct
	var err error
	if len(from.Metadata) > 0 {
		metadata, err = structpb.NewStruct(from.Metadata)
		if err != nil {
			return nil, err
		}
	}

	return &shieldv1beta1.Resource{
		Id:        from.ID,
		Urn:       from.URN,
		Name:      from.Name,
		ProjectId: from.ProjectID,
		Namespace: from.NamespaceID,
		UserId:    from.UserID,
		Metadata:  metadata,
		CreatedAt: timestamppb.New(from.CreatedAt),
		UpdatedAt: timestamppb.New(from.UpdatedAt),
	}, nil
}
