package v1beta1

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
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
	logger := grpczap.Extract(ctx)
	var namespaces []*frontierv1beta1.Namespace

	nsList, err := h.namespaceService.List(ctx)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	for _, ns := range nsList {
		nsPB, err := transformNamespaceToPB(ns)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}

		namespaces = append(namespaces, &nsPB)
	}

	return &frontierv1beta1.ListNamespacesResponse{Namespaces: namespaces}, nil
}

func (h Handler) GetNamespace(ctx context.Context, request *frontierv1beta1.GetNamespaceRequest) (*frontierv1beta1.GetNamespaceResponse, error) {
	logger := grpczap.Extract(ctx)

	fetchedNS, err := h.namespaceService.Get(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, namespace.ErrNotExist),
			errors.Is(err, namespace.ErrInvalidID):
			return nil, grpcNamespaceNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}

	nsPB, err := transformNamespaceToPB(fetchedNS)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
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
