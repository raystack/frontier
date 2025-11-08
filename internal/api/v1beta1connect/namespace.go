package v1beta1connect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/namespace"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type NamespaceService interface {
	Get(ctx context.Context, id string) (namespace.Namespace, error)
	List(ctx context.Context) ([]namespace.Namespace, error)
	Upsert(ctx context.Context, ns namespace.Namespace) (namespace.Namespace, error)
	Update(ctx context.Context, ns namespace.Namespace) (namespace.Namespace, error)
}

func (h *ConnectHandler) ListNamespaces(ctx context.Context, request *connect.Request[frontierv1beta1.ListNamespacesRequest]) (*connect.Response[frontierv1beta1.ListNamespacesResponse], error) {
	errorLogger := NewErrorLogger()

	var namespaces []*frontierv1beta1.Namespace

	nsList, err := h.namespaceService.List(ctx)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "ListNamespaces.List", err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	for _, ns := range nsList {
		nsPB, err := transformNamespaceToPB(ns)
		if err != nil {
			errorLogger.LogTransformError(ctx, request, "ListNamespaces", ns.ID, err)
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}

		namespaces = append(namespaces, &nsPB)
	}

	return connect.NewResponse(&frontierv1beta1.ListNamespacesResponse{Namespaces: namespaces}), nil
}

func (h *ConnectHandler) GetNamespace(ctx context.Context, request *connect.Request[frontierv1beta1.GetNamespaceRequest]) (*connect.Response[frontierv1beta1.GetNamespaceResponse], error) {
	errorLogger := NewErrorLogger()

	fetchedNS, err := h.namespaceService.Get(ctx, request.Msg.GetId())
	if err != nil {
		switch {
		case errors.Is(err, namespace.ErrNotExist),
			errors.Is(err, namespace.ErrInvalidID):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			errorLogger.LogServiceError(ctx, request, "GetNamespace.Get", err,
				zap.String("namespace_id", request.Msg.GetId()))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	nsPB, err := transformNamespaceToPB(fetchedNS)
	if err != nil {
		errorLogger.LogTransformError(ctx, request, "GetNamespace", fetchedNS.ID, err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.GetNamespaceResponse{Namespace: &nsPB}), nil
}

func transformNamespaceToPB(ns namespace.Namespace) (frontierv1beta1.Namespace, error) {
	return frontierv1beta1.Namespace{
		Id:        ns.ID,
		Name:      ns.Name,
		CreatedAt: timestamppb.New(ns.CreatedAt),
		UpdatedAt: timestamppb.New(ns.UpdatedAt),
	}, nil
}
