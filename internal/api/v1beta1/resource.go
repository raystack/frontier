package v1beta1

import (
	"context"
	"errors"

	"github.com/raystack/frontier/core/relation"

	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/internal/bootstrap/schema"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/raystack/frontier/core/resource"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ResourceService interface {
	Get(ctx context.Context, id string) (resource.Resource, error)
	List(ctx context.Context, flt resource.Filter) ([]resource.Resource, error)
	Create(ctx context.Context, resource resource.Resource) (resource.Resource, error)
	Update(ctx context.Context, resource resource.Resource) (resource.Resource, error)
	Delete(ctx context.Context, namespace, id string) error
	CheckAuthz(ctx context.Context, check resource.Check) (bool, error)
	BatchCheck(ctx context.Context, checks []resource.Check) ([]relation.CheckPair, error)
}

var grpcResourceNotFoundErr = status.Errorf(codes.NotFound, "resource doesn't exist")

func (h Handler) ListResources(ctx context.Context, request *frontierv1beta1.ListResourcesRequest) (*frontierv1beta1.ListResourcesResponse, error) {
	var resources []*frontierv1beta1.Resource
	namespaceID := schema.ParseNamespaceAliasIfRequired(request.GetNamespace())
	filters := resource.Filter{
		NamespaceID: namespaceID,
		ProjectID:   request.GetProjectId(),
	}
	resourcesList, err := h.resourceService.List(ctx, filters)
	if err != nil {
		return nil, err
	}

	for _, r := range resourcesList {
		resourcePB, err := transformResourceToPB(r)
		if err != nil {
			return nil, err
		}
		resources = append(resources, resourcePB)
	}

	return &frontierv1beta1.ListResourcesResponse{
		Resources: resources,
	}, nil
}

func (h Handler) ListProjectResources(ctx context.Context, request *frontierv1beta1.ListProjectResourcesRequest) (*frontierv1beta1.ListProjectResourcesResponse, error) {
	var resources []*frontierv1beta1.Resource
	namespaceID := schema.ParseNamespaceAliasIfRequired(request.GetNamespace())
	filters := resource.Filter{
		NamespaceID: namespaceID,
		ProjectID:   request.GetProjectId(),
	}
	resourcesList, err := h.resourceService.List(ctx, filters)
	if err != nil {
		return nil, err
	}

	for _, r := range resourcesList {
		resourcePB, err := transformResourceToPB(r)
		if err != nil {
			return nil, err
		}
		resources = append(resources, resourcePB)
	}

	return &frontierv1beta1.ListProjectResourcesResponse{
		Resources: resources,
	}, nil
}

func (h Handler) CreateProjectResource(ctx context.Context, request *frontierv1beta1.CreateProjectResourceRequest) (*frontierv1beta1.CreateProjectResourceResponse, error) {
	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	var metaDataMap metadata.Metadata
	var err error
	if request.GetBody().GetMetadata() != nil {
		metaDataMap = metadata.Build(request.GetBody().GetMetadata().AsMap())
	}

	parentProject, err := h.projectService.Get(ctx, request.GetProjectId())
	if err != nil {
		return nil, err
	}

	principalType := schema.UserPrincipal
	principalID := request.GetBody().GetPrincipal()
	if ns, id, err := schema.SplitNamespaceAndResourceID(request.GetBody().GetPrincipal()); err == nil {
		principalType = ns
		principalID = id
	}
	namespaceID := schema.ParseNamespaceAliasIfRequired(request.GetBody().GetNamespace())
	newResource, err := h.resourceService.Create(ctx, resource.Resource{
		ID:            request.GetId(),
		Name:          request.GetBody().GetName(),
		Title:         request.GetBody().GetTitle(),
		ProjectID:     parentProject.ID,
		NamespaceID:   namespaceID,
		PrincipalID:   principalID,
		PrincipalType: principalType,
		Metadata:      metaDataMap,
	})
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidEmail):
			return nil, grpcUnauthenticated
		case errors.Is(err, resource.ErrInvalidUUID),
			errors.Is(err, resource.ErrInvalidDetail):
			return nil, grpcBadBodyError
		case errors.Is(err, resource.ErrConflict):
			return nil, grpcConflictError
		default:
			return nil, err
		}
	}

	resourcePB, err := transformResourceToPB(newResource)
	if err != nil {
		return nil, err
	}

	audit.GetAuditor(ctx, parentProject.Organization.ID).Log(audit.ResourceCreatedEvent, audit.Target{
		ID:   newResource.ID,
		Type: newResource.NamespaceID,
	})
	return &frontierv1beta1.CreateProjectResourceResponse{
		Resource: resourcePB,
	}, nil
}

func (h Handler) GetProjectResource(ctx context.Context, request *frontierv1beta1.GetProjectResourceRequest) (*frontierv1beta1.GetProjectResourceResponse, error) {
	fetchedResource, err := h.resourceService.Get(ctx, request.GetId())
	if err != nil {
		switch {
		case errors.Is(err, resource.ErrNotExist),
			errors.Is(err, resource.ErrInvalidUUID),
			errors.Is(err, resource.ErrInvalidID):
			return nil, grpcResourceNotFoundErr
		default:
			return nil, err
		}
	}

	resourcePB, err := transformResourceToPB(fetchedResource)
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.GetProjectResourceResponse{
		Resource: resourcePB,
	}, nil
}

func (h Handler) UpdateProjectResource(ctx context.Context, request *frontierv1beta1.UpdateProjectResourceRequest) (*frontierv1beta1.UpdateProjectResourceResponse, error) {
	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	var metaDataMap metadata.Metadata
	var err error
	if request.GetBody().GetMetadata() != nil {
		metaDataMap = metadata.Build(request.GetBody().GetMetadata().AsMap())
	}

	parentProject, err := h.projectService.Get(ctx, request.GetProjectId())
	if err != nil {
		return nil, err
	}

	principalType := schema.UserPrincipal
	principalID := request.GetBody().GetPrincipal()
	if ns, id, err := schema.SplitNamespaceAndResourceID(request.GetBody().GetPrincipal()); err == nil {
		principalType = ns
		principalID = id
	}
	namespaceID := schema.ParseNamespaceAliasIfRequired(request.GetBody().GetNamespace())
	updatedResource, err := h.resourceService.Update(ctx, resource.Resource{
		ID:            request.GetId(),
		ProjectID:     parentProject.ID,
		NamespaceID:   namespaceID,
		Name:          request.GetBody().GetName(),
		PrincipalID:   principalID,
		PrincipalType: principalType,
		Metadata:      metaDataMap,
	})
	if err != nil {
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
			return nil, err
		}
	}

	resourcePB, err := transformResourceToPB(updatedResource)
	if err != nil {
		return nil, err
	}

	audit.GetAuditor(ctx, parentProject.Organization.ID).Log(audit.ResourceUpdatedEvent, audit.Target{
		ID:   updatedResource.ID,
		Type: updatedResource.NamespaceID,
	})
	return &frontierv1beta1.UpdateProjectResourceResponse{
		Resource: resourcePB,
	}, nil
}

func (h Handler) DeleteProjectResource(ctx context.Context,
	request *frontierv1beta1.DeleteProjectResourceRequest) (*frontierv1beta1.DeleteProjectResourceResponse, error) {
	resourceToDel, err := h.resourceService.Get(ctx, request.GetId())
	if err != nil {
		switch {
		case errors.Is(err, resource.ErrNotExist),
			errors.Is(err, resource.ErrInvalidID),
			errors.Is(err, resource.ErrInvalidUUID):
			return nil, grpcResourceNotFoundErr
		default:
			return nil, err
		}
	}

	parentProject, err := h.projectService.Get(ctx, resourceToDel.ProjectID)
	if err != nil {
		return nil, err
	}

	err = h.resourceService.Delete(ctx, resourceToDel.NamespaceID, resourceToDel.ID)
	if err != nil {
		return nil, err
	}

	audit.GetAuditor(ctx, parentProject.Organization.ID).Log(audit.ResourceDeletedEvent, audit.Target{
		ID:   request.GetId(),
		Type: resourceToDel.NamespaceID,
	})
	return &frontierv1beta1.DeleteProjectResourceResponse{}, nil
}

func transformResourceToPB(from resource.Resource) (*frontierv1beta1.Resource, error) {
	var metadata *structpb.Struct
	var err error
	if len(from.Metadata) > 0 {
		metadata, err = structpb.NewStruct(from.Metadata)
		if err != nil {
			return nil, err
		}
	}

	return &frontierv1beta1.Resource{
		Id:        from.ID,
		Urn:       from.URN,
		Name:      from.Name,
		Title:     from.Title,
		ProjectId: from.ProjectID,
		Namespace: from.NamespaceID,
		Principal: schema.JoinNamespaceAndResourceID(from.PrincipalType, from.PrincipalID),
		Metadata:  metadata,
		CreatedAt: timestamppb.New(from.CreatedAt),
		UpdatedAt: timestamppb.New(from.UpdatedAt),
	}, nil
}
