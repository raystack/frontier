package v1beta1

import (
	"context"
	"errors"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/odpf/shield/core/namespace"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type NamespaceService interface {
	Get(ctx context.Context, id string) (namespace.Namespace, error)
	List(ctx context.Context) ([]namespace.Namespace, error)
	Create(ctx context.Context, ns namespace.Namespace) (namespace.Namespace, error)
	Update(ctx context.Context, ns namespace.Namespace) (namespace.Namespace, error)
}

var grpcNamespaceNotFoundErr = status.Errorf(codes.NotFound, "namespace doesn't exist")

func (h Handler) ListNamespaces(ctx context.Context, request *shieldv1beta1.ListNamespacesRequest) (*shieldv1beta1.ListNamespacesResponse, error) {
	logger := grpczap.Extract(ctx)
	var namespaces []*shieldv1beta1.Namespace

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

	return &shieldv1beta1.ListNamespacesResponse{Namespaces: namespaces}, nil
}

func (h Handler) CreateNamespace(ctx context.Context, request *shieldv1beta1.CreateNamespaceRequest) (*shieldv1beta1.CreateNamespaceResponse, error) {
	logger := grpczap.Extract(ctx)

	newNS, err := h.namespaceService.Create(ctx, namespace.Namespace{
		ID:   request.GetBody().Id,
		Name: request.GetBody().Name,
	})

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	nsPB, err := transformNamespaceToPB(newNS)

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.CreateNamespaceResponse{Namespace: &nsPB}, nil
}

func (h Handler) GetNamespace(ctx context.Context, request *shieldv1beta1.GetNamespaceRequest) (*shieldv1beta1.GetNamespaceResponse, error) {
	logger := grpczap.Extract(ctx)

	fetchedNS, err := h.namespaceService.Get(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, namespace.ErrNotExist):
			return nil, grpcNamespaceNotFoundErr
		case errors.Is(err, namespace.ErrInvalidID):
			return nil, grpcBadBodyError
		default:
			return nil, grpcInternalServerError
		}
	}

	nsPB, err := transformNamespaceToPB(fetchedNS)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.GetNamespaceResponse{Namespace: &nsPB}, nil
}

func (h Handler) UpdateNamespace(ctx context.Context, request *shieldv1beta1.UpdateNamespaceRequest) (*shieldv1beta1.UpdateNamespaceResponse, error) {
	logger := grpczap.Extract(ctx)

	updatedNS, err := h.namespaceService.Update(ctx, namespace.Namespace{
		ID:   request.GetBody().Id,
		Name: request.GetBody().Name,
	})

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	nsPB, err := transformNamespaceToPB(updatedNS)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.UpdateNamespaceResponse{Namespace: &nsPB}, nil
}

func transformNamespaceToPB(ns namespace.Namespace) (shieldv1beta1.Namespace, error) {
	return shieldv1beta1.Namespace{
		Id:        ns.ID,
		Name:      ns.Name,
		CreatedAt: timestamppb.New(ns.CreatedAt),
		UpdatedAt: timestamppb.New(ns.UpdatedAt),
	}, nil
}
