package v1beta1connect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/resource"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *ConnectHandler) ListResources(ctx context.Context, request *connect.Request[frontierv1beta1.ListResourcesRequest]) (*connect.Response[frontierv1beta1.ListResourcesResponse], error) {
	errorLogger := NewErrorLogger()
	var resources []*frontierv1beta1.Resource
	namespaceID := schema.ParseNamespaceAliasIfRequired(request.Msg.GetNamespace())
	filters := resource.Filter{
		NamespaceID: namespaceID,
		ProjectID:   request.Msg.GetProjectId(),
	}
	resourcesList, err := h.resourceService.List(ctx, filters)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "ListResources", err,
			zap.String("namespace", request.Msg.GetNamespace()),
			zap.String("project_id", request.Msg.GetProjectId()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	for _, r := range resourcesList {
		resourcePB, err := transformResourceToPB(r)
		if err != nil {
			errorLogger.LogTransformError(ctx, request, "ListResources", r.ID, err)
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		resources = append(resources, resourcePB)
	}

	return connect.NewResponse(&frontierv1beta1.ListResourcesResponse{
		Resources: resources,
	}), nil
}

func (h *ConnectHandler) ListProjectResources(ctx context.Context, request *connect.Request[frontierv1beta1.ListProjectResourcesRequest]) (*connect.Response[frontierv1beta1.ListProjectResourcesResponse], error) {
	errorLogger := NewErrorLogger()
	var resources []*frontierv1beta1.Resource
	namespaceID := schema.ParseNamespaceAliasIfRequired(request.Msg.GetNamespace())
	filters := resource.Filter{
		NamespaceID: namespaceID,
		ProjectID:   request.Msg.GetProjectId(),
	}
	resourcesList, err := h.resourceService.List(ctx, filters)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "ListProjectResources", err,
			zap.String("namespace", request.Msg.GetNamespace()),
			zap.String("project_id", request.Msg.GetProjectId()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	for _, r := range resourcesList {
		resourcePB, err := transformResourceToPB(r)
		if err != nil {
			errorLogger.LogTransformError(ctx, request, "ListProjectResources", r.ID, err)
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		resources = append(resources, resourcePB)
	}
	return connect.NewResponse(&frontierv1beta1.ListProjectResourcesResponse{
		Resources: resources,
	}), nil
}

func (h *ConnectHandler) CreateProjectResource(ctx context.Context, request *connect.Request[frontierv1beta1.CreateProjectResourceRequest]) (*connect.Response[frontierv1beta1.CreateProjectResourceResponse], error) {
	errorLogger := NewErrorLogger()

	if request.Msg.GetBody() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	var metaDataMap metadata.Metadata
	var err error
	if request.Msg.GetBody().GetMetadata() != nil {
		metaDataMap = metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())
	}

	parentProject, err := h.projectService.Get(ctx, request.Msg.GetProjectId())
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "CreateProjectResource.GetProject", err,
			zap.String("project_id", request.Msg.GetProjectId()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	principalType := schema.UserPrincipal
	principalID := request.Msg.GetBody().GetPrincipal()
	if ns, id, err := schema.SplitNamespaceAndResourceID(request.Msg.GetBody().GetPrincipal()); err == nil {
		principalType = ns
		principalID = id
	}
	namespaceID := schema.ParseNamespaceAliasIfRequired(request.Msg.GetBody().GetNamespace())
	newResource, err := h.resourceService.Create(ctx, resource.Resource{
		ID:            request.Msg.GetId(),
		Name:          request.Msg.GetBody().GetName(),
		Title:         request.Msg.GetBody().GetTitle(),
		ProjectID:     parentProject.ID,
		NamespaceID:   namespaceID,
		PrincipalID:   principalID,
		PrincipalType: principalType,
		Metadata:      metaDataMap,
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "CreateProjectResource", err,
			zap.String("resource_id", request.Msg.GetId()),
			zap.String("project_id", request.Msg.GetProjectId()),
			zap.String("resource_name", request.Msg.GetBody().GetName()),
			zap.String("namespace", request.Msg.GetBody().GetNamespace()),
			zap.String("principal", request.Msg.GetBody().GetPrincipal()))

		switch {
		case errors.Is(err, user.ErrInvalidEmail):
			return nil, connect.NewError(connect.CodeUnauthenticated, ErrUnauthenticated)
		case errors.Is(err, resource.ErrInvalidUUID),
			errors.Is(err, resource.ErrInvalidDetail):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		case errors.Is(err, resource.ErrConflict):
			return nil, connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest)
		default:
			errorLogger.LogUnexpectedError(ctx, request, "CreateProjectResource", err,
				zap.String("resource_id", request.Msg.GetId()),
				zap.String("project_id", request.Msg.GetProjectId()),
				zap.String("resource_name", request.Msg.GetBody().GetName()),
				zap.String("namespace", request.Msg.GetBody().GetNamespace()),
				zap.String("principal", request.Msg.GetBody().GetPrincipal()))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	resourcePB, err := transformResourceToPB(newResource)
	if err != nil {
		errorLogger.LogTransformError(ctx, request, "CreateProjectResource", newResource.ID, err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	audit.GetAuditor(ctx, parentProject.Organization.ID).Log(audit.ResourceCreatedEvent, audit.Target{
		ID:   newResource.ID,
		Type: newResource.NamespaceID,
	})
	return connect.NewResponse(&frontierv1beta1.CreateProjectResourceResponse{
		Resource: resourcePB,
	}), nil
}

func (h *ConnectHandler) GetProjectResource(ctx context.Context, request *connect.Request[frontierv1beta1.GetProjectResourceRequest]) (*connect.Response[frontierv1beta1.GetProjectResourceResponse], error) {
	errorLogger := NewErrorLogger()
	resourceID := request.Msg.GetId()

	fetchedResource, err := h.resourceService.Get(ctx, resourceID)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "GetProjectResource", err,
			zap.String("resource_id", resourceID))

		switch {
		case errors.Is(err, resource.ErrNotExist),
			errors.Is(err, resource.ErrInvalidUUID),
			errors.Is(err, resource.ErrInvalidID):
			return nil, connect.NewError(connect.CodeNotFound, ErrResourceNotFound)
		default:
			errorLogger.LogUnexpectedError(ctx, request, "GetProjectResource", err,
				zap.String("resource_id", resourceID))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	resourcePB, err := transformResourceToPB(fetchedResource)
	if err != nil {
		errorLogger.LogTransformError(ctx, request, "GetProjectResource", fetchedResource.ID, err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.GetProjectResourceResponse{
		Resource: resourcePB,
	}), nil
}

func (h *ConnectHandler) UpdateProjectResource(ctx context.Context, request *connect.Request[frontierv1beta1.UpdateProjectResourceRequest]) (*connect.Response[frontierv1beta1.UpdateProjectResourceResponse], error) {
	errorLogger := NewErrorLogger()

	if request.Msg.GetBody() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	var metaDataMap metadata.Metadata
	var err error
	if request.Msg.GetBody().GetMetadata() != nil {
		metaDataMap = metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())
	}

	parentProject, err := h.projectService.Get(ctx, request.Msg.GetProjectId())
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "UpdateProjectResource.GetProject", err,
			zap.String("project_id", request.Msg.GetProjectId()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	principalType := schema.UserPrincipal
	principalID := request.Msg.GetBody().GetPrincipal()
	if ns, id, err := schema.SplitNamespaceAndResourceID(request.Msg.GetBody().GetPrincipal()); err == nil {
		principalType = ns
		principalID = id
	}
	namespaceID := schema.ParseNamespaceAliasIfRequired(request.Msg.GetBody().GetNamespace())
	updatedResource, err := h.resourceService.Update(ctx, resource.Resource{
		ID:            request.Msg.GetId(),
		ProjectID:     parentProject.ID,
		NamespaceID:   namespaceID,
		Name:          request.Msg.GetBody().GetName(),
		PrincipalID:   principalID,
		PrincipalType: principalType,
		Metadata:      metaDataMap,
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "UpdateProjectResource", err,
			zap.String("resource_id", request.Msg.GetId()),
			zap.String("project_id", request.Msg.GetProjectId()),
			zap.String("resource_name", request.Msg.GetBody().GetName()),
			zap.String("namespace", request.Msg.GetBody().GetNamespace()),
			zap.String("principal", request.Msg.GetBody().GetPrincipal()))

		switch {
		case errors.Is(err, resource.ErrNotExist),
			errors.Is(err, resource.ErrInvalidUUID),
			errors.Is(err, resource.ErrInvalidID):
			return nil, connect.NewError(connect.CodeNotFound, ErrResourceNotFound)
		case errors.Is(err, resource.ErrInvalidDetail),
			errors.Is(err, resource.ErrInvalidURN):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		case errors.Is(err, resource.ErrConflict):
			return nil, connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest)
		default:
			errorLogger.LogUnexpectedError(ctx, request, "UpdateProjectResource", err,
				zap.String("resource_id", request.Msg.GetId()),
				zap.String("project_id", request.Msg.GetProjectId()),
				zap.String("resource_name", request.Msg.GetBody().GetName()),
				zap.String("namespace", request.Msg.GetBody().GetNamespace()),
				zap.String("principal", request.Msg.GetBody().GetPrincipal()))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	resourcePB, err := transformResourceToPB(updatedResource)
	if err != nil {
		errorLogger.LogTransformError(ctx, request, "UpdateProjectResource", updatedResource.ID, err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	audit.GetAuditor(ctx, parentProject.Organization.ID).Log(audit.ResourceUpdatedEvent, audit.Target{
		ID:   updatedResource.ID,
		Type: updatedResource.NamespaceID,
	})
	return connect.NewResponse(&frontierv1beta1.UpdateProjectResourceResponse{
		Resource: resourcePB,
	}), nil
}

func (h *ConnectHandler) DeleteProjectResource(ctx context.Context, request *connect.Request[frontierv1beta1.DeleteProjectResourceRequest]) (*connect.Response[frontierv1beta1.DeleteProjectResourceResponse], error) {
	errorLogger := NewErrorLogger()
	resourceID := request.Msg.GetId()

	resourceToDel, err := h.resourceService.Get(ctx, resourceID)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "DeleteProjectResource.GetResource", err,
			zap.String("resource_id", resourceID))

		switch {
		case errors.Is(err, resource.ErrNotExist),
			errors.Is(err, resource.ErrInvalidID),
			errors.Is(err, resource.ErrInvalidUUID):
			return nil, connect.NewError(connect.CodeNotFound, ErrResourceNotFound)
		default:
			errorLogger.LogUnexpectedError(ctx, request, "DeleteProjectResource.GetResource", err,
				zap.String("resource_id", resourceID))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	parentProject, err := h.projectService.Get(ctx, resourceToDel.ProjectID)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "DeleteProjectResource.GetProject", err,
			zap.String("resource_id", resourceID),
			zap.String("project_id", resourceToDel.ProjectID))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	err = h.resourceService.Delete(ctx, resourceToDel.NamespaceID, resourceToDel.ID)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "DeleteProjectResource", err,
			zap.String("resource_id", resourceID),
			zap.String("project_id", resourceToDel.ProjectID),
			zap.String("namespace", resourceToDel.NamespaceID))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	audit.GetAuditor(ctx, parentProject.Organization.ID).Log(audit.ResourceDeletedEvent, audit.Target{
		ID:   resourceID,
		Type: resourceToDel.NamespaceID,
	})
	return connect.NewResponse(&frontierv1beta1.DeleteProjectResourceResponse{}), nil
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
