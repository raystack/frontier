package v1beta1

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/shield/core/metaschema"
	"github.com/raystack/shield/pkg/metadata"
	shieldv1beta1 "github.com/raystack/shield/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	userMetaSchema  = "user"
	groupMetaSchema = "group"
	orgMetaSchema   = "organization"
	roleMetaSchema  = "role"

	grpcMetaSchemaNotFoundErr = status.Errorf(codes.NotFound, "metaschema doesn't exist")
)

//go:generate mockery --name=MetaSchemaService -r --case underscore --with-expecter --structname MetaSchemaService --filename metaschema_service.go --output=./mocks
type MetaSchemaService interface {
	Get(ctx context.Context, id string) (metaschema.MetaSchema, error)
	Create(ctx context.Context, toCreate metaschema.MetaSchema) (metaschema.MetaSchema, error)
	List(ctx context.Context) ([]metaschema.MetaSchema, error)
	Update(ctx context.Context, id string, toUpdate metaschema.MetaSchema) (metaschema.MetaSchema, error)
	Delete(ctx context.Context, id string) error
	Validate(schema metadata.Metadata, data string) error
}

func (h Handler) ListMetaSchemas(ctx context.Context, request *shieldv1beta1.ListMetaSchemasRequest) (*shieldv1beta1.ListMetaSchemasResponse, error) {
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

	return &shieldv1beta1.ListMetaSchemasResponse{Metaschemas: metaschemas}, nil
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
			errors.Is(err, metaschema.ErrInvalidID),
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

	id := request.GetId()
	if strings.TrimSpace(id) == "" {
		return nil, grpcMetaSchemaNotFoundErr
	}

	fetchedMetaSchema, err := h.metaSchemaService.Get(ctx, id)
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, metaschema.ErrNotExist), errors.Is(err, metaschema.ErrInvalidID):
			return nil, grpcMetaSchemaNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}

	metaschemaPB := transformMetaSchemaToPB(fetchedMetaSchema)

	return &shieldv1beta1.GetMetaSchemaResponse{Metaschema: &metaschemaPB}, nil
}

func (h Handler) UpdateMetaSchema(ctx context.Context, request *shieldv1beta1.UpdateMetaSchemaRequest) (*shieldv1beta1.UpdateMetaSchemaResponse, error) {
	logger := grpczap.Extract(ctx)

	id := request.GetId()
	if strings.TrimSpace(id) == "" {
		return nil, grpcMetaSchemaNotFoundErr
	}
	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	updateMetaSchema, err := h.metaSchemaService.Update(ctx, id, metaschema.MetaSchema{
		Name:   request.GetBody().GetName(),
		Schema: request.GetBody().GetSchema()})

	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, metaschema.ErrInvalidDetail):
			return nil, grpcBadBodyError
		case errors.Is(err, metaschema.ErrInvalidID),
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

	id := request.GetId()
	if strings.TrimSpace(id) == "" {
		return nil, grpcMetaSchemaNotFoundErr
	}

	err := h.metaSchemaService.Delete(ctx, id)
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, metaschema.ErrInvalidID),
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
		Id:        from.ID,
		Name:      from.Name,
		Schema:    from.Schema,
		CreatedAt: timestamppb.New(from.CreatedAt),
		UpdatedAt: timestamppb.New(from.UpdatedAt),
	}
}
