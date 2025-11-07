package v1beta1connect

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/metaschema"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	roleMetaSchema     = "role"
	prospectMetaSchema = "prospect"
	userMetaSchema     = "user"
	groupMetaSchema    = "group"
)

type MetaSchemaService interface {
	Get(ctx context.Context, id string) (metaschema.MetaSchema, error)
	Create(ctx context.Context, toCreate metaschema.MetaSchema) (metaschema.MetaSchema, error)
	List(ctx context.Context) ([]metaschema.MetaSchema, error)
	Update(ctx context.Context, id string, toUpdate metaschema.MetaSchema) (metaschema.MetaSchema, error)
	Delete(ctx context.Context, id string) error
}

func (h *ConnectHandler) ListMetaSchemas(ctx context.Context, req *connect.Request[frontierv1beta1.ListMetaSchemasRequest]) (*connect.Response[frontierv1beta1.ListMetaSchemasResponse], error) {
	errorLogger := NewErrorLogger()

	metaschemasList, err := h.metaSchemaService.List(ctx)
	if err != nil {
		errorLogger.LogServiceError(ctx, req, "ListMetaSchemas.List", err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	var metaschemas []*frontierv1beta1.MetaSchema
	for _, m := range metaschemasList {
		metaschemaPB := transformMetaSchemaToPB(m)
		metaschemas = append(metaschemas, &metaschemaPB)
	}

	return connect.NewResponse(&frontierv1beta1.ListMetaSchemasResponse{
		Metaschemas: metaschemas,
	}), nil
}

func (h *ConnectHandler) CreateMetaSchema(ctx context.Context, req *connect.Request[frontierv1beta1.CreateMetaSchemaRequest]) (*connect.Response[frontierv1beta1.CreateMetaSchemaResponse], error) {
	errorLogger := NewErrorLogger()

	if req.Msg.GetBody() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	newMetaSchema, err := h.metaSchemaService.Create(ctx, metaschema.MetaSchema{
		Name:   req.Msg.GetBody().GetName(),
		Schema: req.Msg.GetBody().GetSchema(),
	})
	if err != nil {
		switch {
		case errors.Is(err, metaschema.ErrNotExist),
			errors.Is(err, metaschema.ErrInvalidID),
			errors.Is(err, metaschema.ErrInvalidDetail):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		case errors.Is(err, metaschema.ErrConflict):
			return nil, connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest)
		default:
			errorLogger.LogServiceError(ctx, req, "CreateMetaSchema.Create", err,
				zap.String("metaschema_name", req.Msg.GetBody().GetName()))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	metaschemaPB := transformMetaSchemaToPB(newMetaSchema)
	return connect.NewResponse(&frontierv1beta1.CreateMetaSchemaResponse{
		Metaschema: &metaschemaPB,
	}), nil
}

func (h *ConnectHandler) GetMetaSchema(ctx context.Context, req *connect.Request[frontierv1beta1.GetMetaSchemaRequest]) (*connect.Response[frontierv1beta1.GetMetaSchemaResponse], error) {
	errorLogger := NewErrorLogger()

	id := req.Msg.GetId()
	if strings.TrimSpace(id) == "" {
		return nil, connect.NewError(connect.CodeNotFound, ErrMetaschemaNotFound)
	}

	fetchedMetaSchema, err := h.metaSchemaService.Get(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, metaschema.ErrNotExist), errors.Is(err, metaschema.ErrInvalidID):
			return nil, connect.NewError(connect.CodeNotFound, ErrMetaschemaNotFound)
		default:
			errorLogger.LogServiceError(ctx, req, "GetMetaSchema.Get", err,
				zap.String("metaschema_id", id))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	metaschemaPB := transformMetaSchemaToPB(fetchedMetaSchema)
	return connect.NewResponse(&frontierv1beta1.GetMetaSchemaResponse{
		Metaschema: &metaschemaPB,
	}), nil
}

func (h *ConnectHandler) UpdateMetaSchema(ctx context.Context, req *connect.Request[frontierv1beta1.UpdateMetaSchemaRequest]) (*connect.Response[frontierv1beta1.UpdateMetaSchemaResponse], error) {
	errorLogger := NewErrorLogger()

	id := req.Msg.GetId()
	if strings.TrimSpace(id) == "" {
		return nil, connect.NewError(connect.CodeNotFound, ErrMetaschemaNotFound)
	}
	if req.Msg.GetBody() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	updateMetaSchema, err := h.metaSchemaService.Update(ctx, id, metaschema.MetaSchema{
		Name:   req.Msg.GetBody().GetName(),
		Schema: req.Msg.GetBody().GetSchema(),
	})
	if err != nil {
		switch {
		case errors.Is(err, metaschema.ErrInvalidDetail):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		case errors.Is(err, metaschema.ErrInvalidID),
			errors.Is(err, metaschema.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrMetaschemaNotFound)
		case errors.Is(err, metaschema.ErrConflict):
			return nil, connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest)
		default:
			errorLogger.LogServiceError(ctx, req, "UpdateMetaSchema.Update", err,
				zap.String("metaschema_id", id),
				zap.String("metaschema_name", req.Msg.GetBody().GetName()))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	metaschemaPB := transformMetaSchemaToPB(updateMetaSchema)
	return connect.NewResponse(&frontierv1beta1.UpdateMetaSchemaResponse{
		Metaschema: &metaschemaPB,
	}), nil
}

func (h *ConnectHandler) DeleteMetaSchema(ctx context.Context, req *connect.Request[frontierv1beta1.DeleteMetaSchemaRequest]) (*connect.Response[frontierv1beta1.DeleteMetaSchemaResponse], error) {
	errorLogger := NewErrorLogger()

	id := req.Msg.GetId()
	if strings.TrimSpace(id) == "" {
		return nil, connect.NewError(connect.CodeNotFound, ErrMetaschemaNotFound)
	}

	err := h.metaSchemaService.Delete(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, metaschema.ErrInvalidID),
			errors.Is(err, metaschema.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrMetaschemaNotFound)
		default:
			errorLogger.LogServiceError(ctx, req, "DeleteMetaSchema.Delete", err,
				zap.String("metaschema_id", id))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	return connect.NewResponse(&frontierv1beta1.DeleteMetaSchemaResponse{}), nil
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
