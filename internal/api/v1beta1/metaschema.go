package v1beta1

import (
	"context"
	"errors"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/odpf/shield/core/metaschema"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var grpcMetaSchemaNotFoundErr = status.Errorf(codes.NotFound, "metaschema doesn't exist")

type MetaSchemaService interface {
	Get(ctx context.Context, name string) (metaschema.MetaSchema, error)
	Create(ctx context.Context, toCreate metaschema.MetaSchema) (metaschema.MetaSchema, error)
	List(ctx context.Context) ([]metaschema.MetaSchema, error)
	Update(ctx context.Context, name string, toUpdate metaschema.MetaSchema) (metaschema.MetaSchema, error)
	Delete(ctx context.Context, name string) error
}

func (h Handler) ListMetaSchema(ctx context.Context, request *shieldv1beta1.ListMetaSchemaRequest) (*shieldv1beta1.ListMetaSchemaResponse, error) {
	logger := grpczap.Extract(ctx)
	var metaschemas []*shieldv1beta1.MetaSchema

	metaschemasList, err := h.metaSchemaService.List(ctx)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	for _, m := range metaschemasList {
		metaschemaPB := transformMetaSchemaToPB(m)
		metaschemas = append(metaschemas, &metaschemaPB)
	}

	return &shieldv1beta1.ListMetaSchemaResponse{Metaschema: metaschemas}, nil
}

func (h Handler) CreateMetaSchema(ctx context.Context, request *shieldv1beta1.CreateMetaSchemaRequest) (*shieldv1beta1.CreateMetaSchemaResponse, error) {
	logger := grpczap.Extract(ctx)

	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	newMetaSchema, err := h.metaSchemaService.Create(ctx, metaschema.MetaSchema{
		Name:   request.GetBody().GetName(),
		Schema: request.GetBody().GetSchema(),
	})
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, metaschema.ErrNotExist),
			errors.Is(err, metaschema.ErrInvalidName),
			errors.Is(err, metaschema.ErrInvalidDetail):
			return nil, grpcBadBodyError
		case errors.Is(err, metaschema.ErrConflict):
			return nil, grpcConflictError
		default:
			return nil, grpcInternalServerError
		}
	}

	metaschemaPB := transformMetaSchemaToPB(newMetaSchema)

	return &shieldv1beta1.CreateMetaSchemaResponse{Metaschema: &metaschemaPB}, nil
}

func (h Handler) GetMetaSchema(ctx context.Context, request *shieldv1beta1.GetMetaSchemaRequest) (*shieldv1beta1.GetMetaSchemaResponse, error) {
	logger := grpczap.Extract(ctx)

	fetchedMetaSchema, err := h.metaSchemaService.Get(ctx, request.GetName())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, metaschema.ErrNotExist), errors.Is(err, metaschema.ErrInvalidName):
			return nil, grpcRoleNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}

	metaschemaPB := transformMetaSchemaToPB(fetchedMetaSchema)

	return &shieldv1beta1.GetMetaSchemaResponse{Metaschema: &metaschemaPB}, nil
}

func (h Handler) UpdateMetaSchema(ctx context.Context, request *shieldv1beta1.UpdateMetaSchemaRequest) (*shieldv1beta1.UpdateMetaSchemaResponse, error) {
	logger := grpczap.Extract(ctx)

	// todo  validate name in proto
	name := request.GetName()

	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	if name != request.GetBody().GetName() {
		return nil, status.Errorf(codes.InvalidArgument, ErrBadRequest.Error())
	}

	updateMetaSchema, err := h.metaSchemaService.Update(ctx, name, metaschema.MetaSchema{
		Name:   request.GetBody().GetName(),
		Schema: request.GetBody().GetSchema()})

	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, metaschema.ErrInvalidDetail):
			return nil, grpcBadBodyError
		case errors.Is(err, metaschema.ErrInvalidName),
			errors.Is(err, metaschema.ErrNotExist):
			return nil, grpcMetaSchemaNotFoundErr
		case errors.Is(err, metaschema.ErrConflict):
			return nil, grpcConflictError
		default:
			return nil, grpcInternalServerError
		}
	}

	metaschemaPB := transformMetaSchemaToPB(updateMetaSchema)
	return &shieldv1beta1.UpdateMetaSchemaResponse{Metaschema: &metaschemaPB}, nil
}

func (h Handler) DeleteMetaSchema(ctx context.Context, request *shieldv1beta1.DeleteMetaSchemaRequest) (*shieldv1beta1.DeleteMetaSchemaResponse, error) {
	logger := grpczap.Extract(ctx)

	err := h.metaSchemaService.Delete(ctx, request.GetName())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, metaschema.ErrInvalidName),
			errors.Is(err, metaschema.ErrNotExist):
			return nil, grpcMetaSchemaNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}

	return &shieldv1beta1.DeleteMetaSchemaResponse{}, nil
}

func transformMetaSchemaToPB(from metaschema.MetaSchema) shieldv1beta1.MetaSchema {
	return shieldv1beta1.MetaSchema{
		Name:   from.Name,
		Schema: from.Schema,
	}
}
