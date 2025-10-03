package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/resource"
	"github.com/raystack/frontier/internal/bootstrap/schema"
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
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	for _, r := range resourcesList {
		resourcePB, err := transformResourceToPB(r)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
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
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	for _, r := range resourcesList {
		resourcePB, err := transformResourceToPB(r)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		resources = append(resources, resourcePB)
	}
	return connect.NewResponse(&frontierv1beta1.ListProjectResourcesResponse{
		Resources: resources,
	}), nil
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
