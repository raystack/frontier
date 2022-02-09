package v1beta1

import (
	"context"
	"errors"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/odpf/shield/internal/schema"
	"github.com/odpf/shield/model"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type NamespaceService interface {
	GetNamespace(ctx context.Context, id string) (model.Namespace, error)
	ListNamespaces(ctx context.Context) ([]model.Namespace, error)
	CreateNamespace(ctx context.Context, ns model.Namespace) (model.Namespace, error)
	UpdateNamespace(ctx context.Context, id string, ns model.Namespace) (model.Namespace, error)
}

var grpcNamespaceNotFoundErr = status.Errorf(codes.NotFound, "namespace doesn't exist")

func (v Dep) ListNamespaces(ctx context.Context, request *shieldv1beta1.ListNamespacesRequest) (*shieldv1beta1.ListNamespacesResponse, error) {
	logger := grpczap.Extract(ctx)
	var namespaces []*shieldv1beta1.Namespace

	nsList, err := v.NamespaceService.ListNamespaces(ctx)
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

func (v Dep) CreateNamespace(ctx context.Context, request *shieldv1beta1.CreateNamespaceRequest) (*shieldv1beta1.CreateNamespaceResponse, error) {
	logger := grpczap.Extract(ctx)

	newNS, err := v.NamespaceService.CreateNamespace(ctx, model.Namespace{
		Id:   request.GetBody().Id,
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

func (v Dep) GetNamespace(ctx context.Context, request *shieldv1beta1.GetNamespaceRequest) (*shieldv1beta1.GetNamespaceResponse, error) {
	logger := grpczap.Extract(ctx)

	fetchedNS, err := v.NamespaceService.GetNamespace(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, schema.NamespaceDoesntExist):
			return nil, grpcNamespaceNotFoundErr
		case errors.Is(err, schema.InvalidUUID):
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

func (v Dep) UpdateNamespace(ctx context.Context, request *shieldv1beta1.UpdateNamespaceRequest) (*shieldv1beta1.UpdateNamespaceResponse, error) {
	logger := grpczap.Extract(ctx)

	updatedNS, err := v.NamespaceService.UpdateNamespace(ctx, request.GetId(), model.Namespace{
		Id:   request.GetBody().Id,
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

func transformNamespaceToPB(ns model.Namespace) (shieldv1beta1.Namespace, error) {
	return shieldv1beta1.Namespace{
		Id:        ns.Id,
		Name:      ns.Name,
		CreatedAt: timestamppb.New(ns.CreatedAt),
		UpdatedAt: timestamppb.New(ns.UpdatedAt),
	}, nil
}
