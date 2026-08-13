package reconcile

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/metaschema"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

// MetaSchemaAPI is the API subset the metaschema reconciler needs. Every call
// lives on FrontierService; the caller provides one value that serves it. There
// is no delete: built-in metaschemas are values, reset but never removed.
type MetaSchemaAPI interface {
	ListMetaSchemas(context.Context, *connect.Request[frontierv1beta1.ListMetaSchemasRequest]) (*connect.Response[frontierv1beta1.ListMetaSchemasResponse], error)
	CreateMetaSchema(context.Context, *connect.Request[frontierv1beta1.CreateMetaSchemaRequest]) (*connect.Response[frontierv1beta1.CreateMetaSchemaResponse], error)
	UpdateMetaSchema(context.Context, *connect.Request[frontierv1beta1.UpdateMetaSchemaRequest]) (*connect.Response[frontierv1beta1.UpdateMetaSchemaResponse], error)
}

// MetaSchemaReconciler makes the built-in metaschemas match the desired spec. The
// name is the identity; the JSON schema is the managed value. Built-ins are
// values: one left out of the file resets to its shipped default.
type MetaSchemaReconciler struct {
	client   MetaSchemaAPI
	header   string
	defaults map[string]string
}

func NewMetaSchemaReconciler(client MetaSchemaAPI, header string) *MetaSchemaReconciler {
	return &MetaSchemaReconciler{client: client, header: header, defaults: metaschema.Defaults}
}

func (r *MetaSchemaReconciler) Kind() string { return KindMetaSchema }

// Validate checks every entry without touching the server, so a bad entry stops
// the whole file before anything applies.
func (r *MetaSchemaReconciler) Validate(spec []byte) error {
	var specs []MetaSchemaSpec
	if err := decodeSpec(spec, &specs); err != nil {
		return fmt.Errorf("parse %s spec: %w", KindMetaSchema, err)
	}
	return validateMetaSchemaSpecs(specs, r.defaults)
}

func (r *MetaSchemaReconciler) Reconcile(ctx context.Context, spec []byte, dryRun bool) (Report, error) {
	var specs []MetaSchemaSpec
	if err := decodeSpec(spec, &specs); err != nil {
		return Report{}, fmt.Errorf("parse %s spec: %w", KindMetaSchema, err)
	}

	current, err := r.fetchCurrent(ctx)
	if err != nil {
		return Report{}, err
	}

	ops, err := diffMetaSchemas(specs, current, r.defaults)
	if err != nil {
		return Report{}, err
	}

	rep := Report{Kind: KindMetaSchema, DryRun: dryRun}
	for _, op := range ops {
		rep.Planned = append(rep.Planned, op.String())
	}
	if dryRun {
		return rep, nil
	}
	for _, op := range ops {
		if err := r.apply(ctx, op); err != nil {
			return rep, fmt.Errorf("apply [%s]: %w", op, err)
		}
		rep.Applied++
	}
	return rep, nil
}

// Export returns the built-ins whose schema differs from the default as a
// desired-state spec, so reconciling it plans no changes.
func (r *MetaSchemaReconciler) Export(ctx context.Context) (any, error) {
	current, err := r.fetchCurrent(ctx)
	if err != nil {
		return nil, err
	}
	return exportMetaSchemas(current, r.defaults)
}

func (r *MetaSchemaReconciler) fetchCurrent(ctx context.Context) ([]currentMetaSchema, error) {
	resp, err := r.client.ListMetaSchemas(ctx, authReq(&frontierv1beta1.ListMetaSchemasRequest{}, r.header))
	if err != nil {
		return nil, fmt.Errorf("list metaschemas: %w", err)
	}
	var current []currentMetaSchema
	for _, ms := range resp.Msg.GetMetaschemas() {
		current = append(current, currentMetaSchema{
			ID:     ms.GetId(),
			Name:   ms.GetName(),
			Schema: ms.GetSchema(),
		})
	}
	return current, nil
}

func (r *MetaSchemaReconciler) apply(ctx context.Context, op metaSchemaOp) error {
	body := &frontierv1beta1.MetaSchemaRequestBody{Name: op.name, Schema: op.schema}
	if op.id == "" {
		_, err := r.client.CreateMetaSchema(ctx, authReq(&frontierv1beta1.CreateMetaSchemaRequest{Body: body}, r.header))
		return err
	}
	_, err := r.client.UpdateMetaSchema(ctx, authReq(&frontierv1beta1.UpdateMetaSchemaRequest{Id: op.id, Body: body}, r.header))
	return err
}
