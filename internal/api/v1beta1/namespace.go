package v1beta1

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/raystack/frontier/core/namespace"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type NamespaceService interface {
	Get(ctx context.Context, id string) (namespace.Namespace, error)
	List(ctx context.Context) ([]namespace.Namespace, error)
	Upsert(ctx context.Context, ns namespace.Namespace) (namespace.Namespace, error)
	Update(ctx context.Context, ns namespace.Namespace) (namespace.Namespace, error)
}

var grpcNamespaceNotFoundErr = status.Errorf(codes.NotFound, "namespace doesn't exist")

func (h Handler) ListNamespaces(ctx context.Context, request *frontierv1beta1.ListNamespacesRequest) (*frontierv1beta1.ListNamespacesResponse, error) {
	var namespaces []*frontierv1beta1.Namespace

	nsList, err := h.namespaceService.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, ns := range nsList {
		nsPB, err := transformNamespaceToPB(ns)
		if err != nil {
			return nil, err
		}

		namespaces = append(namespaces, &nsPB)
	}

	return &frontierv1beta1.ListNamespacesResponse{Namespaces: namespaces}, nil
}

func (h Handler) GetNamespace(ctx context.Context, request *frontierv1beta1.GetNamespaceRequest) (*frontierv1beta1.GetNamespaceResponse, error) {
	fetchedNS, err := h.namespaceService.Get(ctx, request.GetId())
	if err != nil {
		switch {
		case errors.Is(err, namespace.ErrNotExist),
			errors.Is(err, namespace.ErrInvalidID):
			return nil, grpcNamespaceNotFoundErr
		default:
			return nil, err
		}
	}

	nsPB, err := transformNamespaceToPB(fetchedNS)
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.GetNamespaceResponse{Namespace: &nsPB}, nil
}

func transformNamespaceToPB(ns namespace.Namespace) (frontierv1beta1.Namespace, error) {
	return frontierv1beta1.Namespace{
		Id:        ns.ID,
		Name:      ns.Name,
		CreatedAt: timestamppb.New(ns.CreatedAt),
		UpdatedAt: timestamppb.New(ns.UpdatedAt),
	}, nil
}
