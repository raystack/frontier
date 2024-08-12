package v1beta1

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"github.com/raystack/frontier/core/metaschema"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
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

type MetaSchemaService interface {
	Get(ctx context.Context, id string) (metaschema.MetaSchema, error)
	Create(ctx context.Context, toCreate metaschema.MetaSchema) (metaschema.MetaSchema, error)
	List(ctx context.Context) ([]metaschema.MetaSchema, error)
	Update(ctx context.Context, id string, toUpdate metaschema.MetaSchema) (metaschema.MetaSchema, error)
	Delete(ctx context.Context, id string) error
	Validate(schema metadata.Metadata, data string) error
}

func (h Handler) ListMetaSchemas(ctx context.Context, request *frontierv1beta1.ListMetaSchemasRequest) (*frontierv1beta1.ListMetaSchemasResponse, error) {
	var metaschemas []*frontierv1beta1.MetaSchema

	metaschemasList, err := h.metaSchemaService.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, m := range metaschemasList {
		metaschemaPB := transformMetaSchemaToPB(m)
		metaschemas = append(metaschemas, &metaschemaPB)
	}

	return &frontierv1beta1.ListMetaSchemasResponse{Metaschemas: metaschemas}, nil
}

func (h Handler) CreateMetaSchema(ctx context.Context, request *frontierv1beta1.CreateMetaSchemaRequest) (*frontierv1beta1.CreateMetaSchemaResponse, error) {
	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	newMetaSchema, err := h.metaSchemaService.Create(ctx, metaschema.MetaSchema{
		Name:   request.GetBody().GetName(),
		Schema: request.GetBody().GetSchema(),
	})
	if err != nil {
		switch {
		case errors.Is(err, metaschema.ErrNotExist),
			errors.Is(err, metaschema.ErrInvalidID),
			errors.Is(err, metaschema.ErrInvalidDetail):
			return nil, grpcBadBodyError
		case errors.Is(err, metaschema.ErrConflict):
			return nil, grpcConflictError
		default:
			return nil, err
		}
	}

	metaschemaPB := transformMetaSchemaToPB(newMetaSchema)
	return &frontierv1beta1.CreateMetaSchemaResponse{Metaschema: &metaschemaPB}, nil
}

func (h Handler) GetMetaSchema(ctx context.Context, request *frontierv1beta1.GetMetaSchemaRequest) (*frontierv1beta1.GetMetaSchemaResponse, error) {
	id := request.GetId()
	if strings.TrimSpace(id) == "" {
		return nil, grpcMetaSchemaNotFoundErr
	}

	fetchedMetaSchema, err := h.metaSchemaService.Get(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, metaschema.ErrNotExist), errors.Is(err, metaschema.ErrInvalidID):
			return nil, grpcMetaSchemaNotFoundErr
		default:
			return nil, err
		}
	}

	metaschemaPB := transformMetaSchemaToPB(fetchedMetaSchema)

	return &frontierv1beta1.GetMetaSchemaResponse{Metaschema: &metaschemaPB}, nil
}

func (h Handler) UpdateMetaSchema(ctx context.Context, request *frontierv1beta1.UpdateMetaSchemaRequest) (*frontierv1beta1.UpdateMetaSchemaResponse, error) {
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
		switch {
		case errors.Is(err, metaschema.ErrInvalidDetail):
			return nil, grpcBadBodyError
		case errors.Is(err, metaschema.ErrInvalidID),
			errors.Is(err, metaschema.ErrNotExist):
			return nil, grpcMetaSchemaNotFoundErr
		case errors.Is(err, metaschema.ErrConflict):
			return nil, grpcConflictError
		default:
			return nil, err
		}
	}

	metaschemaPB := transformMetaSchemaToPB(updateMetaSchema)
	return &frontierv1beta1.UpdateMetaSchemaResponse{Metaschema: &metaschemaPB}, nil
}

func (h Handler) DeleteMetaSchema(ctx context.Context, request *frontierv1beta1.DeleteMetaSchemaRequest) (*frontierv1beta1.DeleteMetaSchemaResponse, error) {
	id := request.GetId()
	if strings.TrimSpace(id) == "" {
		return nil, grpcMetaSchemaNotFoundErr
	}

	err := h.metaSchemaService.Delete(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, metaschema.ErrInvalidID),
			errors.Is(err, metaschema.ErrNotExist):
			return nil, grpcMetaSchemaNotFoundErr
		default:
			return nil, err
		}
	}

	return &frontierv1beta1.DeleteMetaSchemaResponse{}, nil
}

func transformMetaSchemaToPB(from metaschema.MetaSchema) frontierv1beta1.MetaSchema {
	return frontierv1beta1.MetaSchema{
		Id:        from.ID,
		Name:      from.Name,
		Schema:    from.Schema,
		CreatedAt: timestamppb.New(from.CreatedAt),
		UpdatedAt: timestamppb.New(from.UpdatedAt),
	}
}
