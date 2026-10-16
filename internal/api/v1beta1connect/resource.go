package v1beta1connect

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/resource"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *ConnectHandler) ListResources(ctx context.Context, request *connect.Request[frontierv1beta1.ListResourcesRequest]) (*connect.Response[frontierv1beta1.ListResourcesResponse], error) {
	var resources []*frontierv1beta1.Resource
	namespaceID := schema.ParseNamespaceAliasIfRequired(request.Msg.GetNamespace())
	filters := resource.Filter{
		NamespaceID: namespaceID,
		ProjectID:   request.Msg.GetProjectId(),
	}
	resourcesList, err := h.resourceService.List(ctx, filters)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListResources: namespace=%s project_id=%s: %w", request.Msg.GetNamespace(), request.Msg.GetProjectId(), err))
	}

	for _, r := range resourcesList {
		resourcePB, err := transformResourceToPB(r)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListResources: entity_id=%s: %w", r.ID, err))
		}
		resources = append(resources, resourcePB)
	}

	return connect.NewResponse(&frontierv1beta1.ListResourcesResponse{
		Resources: resources,
	}), nil
}

func (h *ConnectHandler) ListProjectResources(ctx context.Context, request *connect.Request[frontierv1beta1.ListProjectResourcesRequest]) (*connect.Response[frontierv1beta1.ListProjectResourcesResponse], error) {
	var resources []*frontierv1beta1.Resource
	namespaceID := schema.ParseNamespaceAliasIfRequired(request.Msg.GetNamespace())
	filters := resource.Filter{
		NamespaceID: namespaceID,
		ProjectID:   request.Msg.GetProjectId(),
	}
	resourcesList, err := h.resourceService.List(ctx, filters)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListProjectResources: namespace=%s project_id=%s: %w", request.Msg.GetNamespace(), request.Msg.GetProjectId(), err))
	}
	for _, r := range resourcesList {
		resourcePB, err := transformResourceToPB(r)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListProjectResources: entity_id=%s: %w", r.ID, err))
		}
		resources = append(resources, resourcePB)
	}
	return connect.NewResponse(&frontierv1beta1.ListProjectResourcesResponse{
		Resources: resources,
	}), nil
}

func (h *ConnectHandler) CreateProjectResource(ctx context.Context, request *connect.Request[frontierv1beta1.CreateProjectResourceRequest]) (*connect.Response[frontierv1beta1.CreateProjectResourceResponse], error) {
	if request.Msg.GetBody() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	resourcePB, err := h.createResource(ctx, request, "CreateProjectResource", request.Msg.GetId(), request.Msg.GetBody(),
		resource.Resource{ProjectID: request.Msg.GetProjectId()}, "project_id", request.Msg.GetProjectId())
	if err != nil {
		return nil, err
	}
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
			"resource_id", resourceID)

		switch {
		case errors.Is(err, resource.ErrNotExist),
			errors.Is(err, resource.ErrInvalidUUID),
			errors.Is(err, resource.ErrInvalidID):
			return nil, connect.NewError(connect.CodeNotFound, ErrResourceNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("GetProjectResource: resource_id=%s: %w", resourceID, err))
		}
	}

	resourcePB, err := transformResourceToPB(fetchedResource)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("GetProjectResource: entity_id=%s: %w", fetchedResource.ID, err))
	}

	return connect.NewResponse(&frontierv1beta1.GetProjectResourceResponse{
		Resource: resourcePB,
	}), nil
}

func (h *ConnectHandler) UpdateProjectResource(ctx context.Context, request *connect.Request[frontierv1beta1.UpdateProjectResourceRequest]) (*connect.Response[frontierv1beta1.UpdateProjectResourceResponse], error) {
	if request.Msg.GetBody() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	parentProject, err := h.projectService.Get(ctx, request.Msg.GetProjectId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("UpdateProjectResource.GetProject: project_id=%s: %w", request.Msg.GetProjectId(), err))
	}

	resourcePB, err := h.updateResource(ctx, request, "UpdateProjectResource", request.Msg.GetId(), request.Msg.GetBody(),
		resource.Resource{ProjectID: parentProject.ID}, "project_id", request.Msg.GetProjectId(), parentProject.Organization.ID)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&frontierv1beta1.UpdateProjectResourceResponse{
		Resource: resourcePB,
	}), nil
}

func (h *ConnectHandler) DeleteProjectResource(ctx context.Context, request *connect.Request[frontierv1beta1.DeleteProjectResourceRequest]) (*connect.Response[frontierv1beta1.DeleteProjectResourceResponse], error) {
	resourceID := request.Msg.GetId()

	resourceToDel, err := h.getResourceToDelete(ctx, request, "DeleteProjectResource", resourceID)
	if err != nil {
		return nil, err
	}

	parentProject, err := h.projectService.Get(ctx, resourceToDel.ProjectID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("DeleteProjectResource.GetProject: resource_id=%s project_id=%s: %w", resourceID, resourceToDel.ProjectID, err))
	}

	if err := h.deleteResource(ctx, request, "DeleteProjectResource", resourceID, resourceToDel,
		"project_id", resourceToDel.ProjectID, parentProject.Organization.ID); err != nil {
		return nil, err
	}
	return connect.NewResponse(&frontierv1beta1.DeleteProjectResourceResponse{}), nil
}

func (h *ConnectHandler) DeleteOrganizationResource(ctx context.Context, request *connect.Request[frontierv1beta1.DeleteOrganizationResourceRequest]) (*connect.Response[frontierv1beta1.DeleteOrganizationResourceResponse], error) {
	resourceID := request.Msg.GetId()

	resourceToDel, err := h.getResourceToDelete(ctx, request, "DeleteOrganizationResource", resourceID)
	if err != nil {
		return nil, err
	}

	if err := h.deleteResource(ctx, request, "DeleteOrganizationResource", resourceID, resourceToDel,
		"org_id", resourceToDel.OrgID, resourceToDel.OrgID); err != nil {
		return nil, err
	}
	return connect.NewResponse(&frontierv1beta1.DeleteOrganizationResourceResponse{}), nil
}

func (h *ConnectHandler) getResourceToDelete(ctx context.Context, request connect.AnyRequest, method, id string) (resource.Resource, error) {
	res, err := h.resourceService.Get(ctx, id)
	if err != nil {
		NewErrorLogger().LogServiceError(ctx, request, method+".GetResource", err,
			"resource_id", id)

		switch {
		case errors.Is(err, resource.ErrNotExist),
			errors.Is(err, resource.ErrInvalidID),
			errors.Is(err, resource.ErrInvalidUUID):
			return resource.Resource{}, connect.NewError(connect.CodeNotFound, ErrResourceNotFound)
		default:
			return resource.Resource{}, connect.NewError(connect.CodeInternal, fmt.Errorf("%s.GetResource: resource_id=%s: %w", method, id, err))
		}
	}
	return res, nil
}

func (h *ConnectHandler) deleteResource(ctx context.Context, request connect.AnyRequest, method, id string,
	res resource.Resource, parentKey, parentValue, auditOrgID string,
) error {
	if err := h.resourceService.Delete(ctx, res.NamespaceID, res.ID); err != nil {
		if errors.Is(err, resource.ErrNotExist) {
			return connect.NewError(connect.CodeNotFound, ErrResourceNotFound)
		}
		return connect.NewError(connect.CodeInternal, fmt.Errorf("%s: resource_id=%s %s=%s namespace=%s: %w", method, id, parentKey, parentValue, res.NamespaceID, err))
	}

	if err := audit.GetAuditor(ctx, auditOrgID).Log(audit.ResourceDeletedEvent, audit.Target{
		ID:   id,
		Type: res.NamespaceID,
	}); err != nil {
		NewErrorLogger().LogServiceError(ctx, request, method+".AuditLog", err,
			"resource_id", id)
	}
	return nil
}

func (h *ConnectHandler) CreateOrganizationResource(ctx context.Context, request *connect.Request[frontierv1beta1.CreateOrganizationResourceRequest]) (*connect.Response[frontierv1beta1.CreateOrganizationResourceResponse], error) {
	if request.Msg.GetBody() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	resourcePB, err := h.createResource(ctx, request, "CreateOrganizationResource", request.Msg.GetId(), request.Msg.GetBody(),
		resource.Resource{OrgID: request.Msg.GetOrgId()}, "org_id", request.Msg.GetOrgId())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&frontierv1beta1.CreateOrganizationResourceResponse{
		Resource: resourcePB,
	}), nil
}

func (h *ConnectHandler) createResource(ctx context.Context, request connect.AnyRequest, method, id string,
	body *frontierv1beta1.ResourceRequestBody, parent resource.Resource, parentKey, parentValue string,
) (*frontierv1beta1.Resource, error) {
	errorLogger := NewErrorLogger()

	newResource, err := h.resourceService.Create(ctx, resourceFromBody(id, body, parent))
	if err != nil {
		errorLogger.LogServiceError(ctx, request, method, err,
			"resource_id", id,
			parentKey, parentValue,
			"resource_name", body.GetName(),
			"namespace", body.GetNamespace(),
			"principal", body.GetPrincipal())

		switch {
		case errors.Is(err, project.ErrNotExist),
			errors.Is(err, project.ErrInvalidID),
			errors.Is(err, project.ErrInvalidUUID):
			return nil, connect.NewError(connect.CodeNotFound, ErrProjectNotFound)
		case errors.Is(err, organization.ErrNotExist),
			errors.Is(err, organization.ErrInvalidID),
			errors.Is(err, organization.ErrInvalidUUID):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgNotFound)
		case errors.Is(err, user.ErrInvalidEmail):
			return nil, connect.NewError(connect.CodeUnauthenticated, ErrUnauthenticated)
		case errors.Is(err, resource.ErrInvalidUUID),
			errors.Is(err, resource.ErrInvalidDetail):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		case errors.Is(err, resource.ErrConflict):
			return nil, connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("%s: resource_id=%s %s=%s resource_name=%s namespace=%s principal=%s: %w", method, id, parentKey, parentValue, body.GetName(), body.GetNamespace(), body.GetPrincipal(), err))
		}
	}

	resourcePB, err := transformResourceToPB(newResource)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("%s: entity_id=%s: %w", method, newResource.ID, err))
	}

	if err := audit.GetAuditor(ctx, newResource.OrgID).Log(audit.ResourceCreatedEvent, audit.Target{
		ID:   newResource.ID,
		Type: newResource.NamespaceID,
	}); err != nil {
		errorLogger.LogServiceError(ctx, request, method+".AuditLog", err,
			"resource_id", newResource.ID)
	}
	return resourcePB, nil
}

func (h *ConnectHandler) UpdateOrganizationResource(ctx context.Context, request *connect.Request[frontierv1beta1.UpdateOrganizationResourceRequest]) (*connect.Response[frontierv1beta1.UpdateOrganizationResourceResponse], error) {
	if request.Msg.GetBody() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	resourcePB, err := h.updateResource(ctx, request, "UpdateOrganizationResource", request.Msg.GetId(), request.Msg.GetBody(),
		resource.Resource{OrgID: request.Msg.GetOrgId()}, "org_id", request.Msg.GetOrgId(), request.Msg.GetOrgId())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&frontierv1beta1.UpdateOrganizationResourceResponse{
		Resource: resourcePB,
	}), nil
}

func (h *ConnectHandler) updateResource(ctx context.Context, request connect.AnyRequest, method, id string,
	body *frontierv1beta1.ResourceRequestBody, parent resource.Resource, parentKey, parentValue, auditOrgID string,
) (*frontierv1beta1.Resource, error) {
	errorLogger := NewErrorLogger()

	updatedResource, err := h.resourceService.Update(ctx, resourceFromBody(id, body, parent))
	if err != nil {
		errorLogger.LogServiceError(ctx, request, method, err,
			"resource_id", id,
			parentKey, parentValue,
			"resource_name", body.GetName(),
			"namespace", body.GetNamespace(),
			"principal", body.GetPrincipal())

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
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("%s: resource_id=%s %s=%s resource_name=%s namespace=%s principal=%s: %w", method, id, parentKey, parentValue, body.GetName(), body.GetNamespace(), body.GetPrincipal(), err))
		}
	}

	resourcePB, err := transformResourceToPB(updatedResource)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("%s: entity_id=%s: %w", method, updatedResource.ID, err))
	}

	if err := audit.GetAuditor(ctx, auditOrgID).Log(audit.ResourceUpdatedEvent, audit.Target{
		ID:   updatedResource.ID,
		Type: updatedResource.NamespaceID,
	}); err != nil {
		errorLogger.LogServiceError(ctx, request, method+".AuditLog", err,
			"resource_id", updatedResource.ID)
	}
	return resourcePB, nil
}

func resourceFromBody(id string, body *frontierv1beta1.ResourceRequestBody, parent resource.Resource) resource.Resource {
	var metaDataMap metadata.Metadata
	if body.GetMetadata() != nil {
		metaDataMap = metadata.Build(body.GetMetadata().AsMap())
	}

	principalType := schema.UserPrincipal
	principalID := body.GetPrincipal()
	if ns, pid, err := schema.SplitNamespaceAndResourceID(body.GetPrincipal()); err == nil {
		principalType = ns
		principalID = pid
	}

	res := parent
	res.ID = id
	res.Name = body.GetName()
	res.Title = body.GetTitle()
	res.NamespaceID = schema.ParseNamespaceAliasIfRequired(body.GetNamespace())
	res.PrincipalID = principalID
	res.PrincipalType = principalType
	res.Metadata = metaDataMap
	return res
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
		OrgId:     from.OrgID,
		Namespace: from.NamespaceID,
		Principal: schema.JoinNamespaceAndResourceID(from.PrincipalType, from.PrincipalID),
		Metadata:  metadata,
		CreatedAt: timestamppb.New(from.CreatedAt),
		UpdatedAt: timestamppb.New(from.UpdatedAt),
	}, nil
}
