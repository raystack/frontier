package reconcile

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/metaschema"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
)

type updateCall struct {
	id     string
	name   string
	schema string
}

type fakeMetaSchemaAPI struct {
	schemas []*frontierv1beta1.MetaSchema
	created []*frontierv1beta1.MetaSchemaRequestBody
	updated []updateCall
}

func (f *fakeMetaSchemaAPI) ListMetaSchemas(_ context.Context, _ *connect.Request[frontierv1beta1.ListMetaSchemasRequest]) (*connect.Response[frontierv1beta1.ListMetaSchemasResponse], error) {
	return connect.NewResponse(&frontierv1beta1.ListMetaSchemasResponse{Metaschemas: f.schemas}), nil
}

func (f *fakeMetaSchemaAPI) CreateMetaSchema(_ context.Context, req *connect.Request[frontierv1beta1.CreateMetaSchemaRequest]) (*connect.Response[frontierv1beta1.CreateMetaSchemaResponse], error) {
	f.created = append(f.created, req.Msg.GetBody())
	return connect.NewResponse(&frontierv1beta1.CreateMetaSchemaResponse{}), nil
}

func (f *fakeMetaSchemaAPI) UpdateMetaSchema(_ context.Context, req *connect.Request[frontierv1beta1.UpdateMetaSchemaRequest]) (*connect.Response[frontierv1beta1.UpdateMetaSchemaResponse], error) {
	f.updated = append(f.updated, updateCall{id: req.Msg.GetId(), name: req.Msg.GetBody().GetName(), schema: req.Msg.GetBody().GetSchema()})
	return connect.NewResponse(&frontierv1beta1.UpdateMetaSchemaResponse{}), nil
}

// seededDefaults returns the five built-ins at their default schema, each with a
// stable fake id of "<name>-id".
func seededDefaults() []*frontierv1beta1.MetaSchema {
	var out []*frontierv1beta1.MetaSchema
	for name, schema := range metaschema.Defaults {
		out = append(out, &frontierv1beta1.MetaSchema{Id: name + "-id", Name: name, Schema: schema})
	}
	return out
}

func TestMetaSchemaReconciler(t *testing.T) {
	t.Run("updates a built-in the file overrides", func(t *testing.T) {
		api := &fakeMetaSchemaAPI{schemas: seededDefaults()}
		spec := []byte("- {name: organization, schema: '{\"type\":\"object\",\"required\":[\"cost_center\"]}'}\n")

		rep, err := NewMetaSchemaReconciler(api, "").Reconcile(context.Background(), spec, false)

		assert.NoError(t, err)
		assert.Equal(t, []string{"set metaschema organization"}, rep.Planned)
		assert.Equal(t, 1, rep.Applied)
		assert.Empty(t, api.created)
		assert.Len(t, api.updated, 1)
		assert.Equal(t, "organization-id", api.updated[0].id)
	})

	t.Run("dry run plans but applies nothing", func(t *testing.T) {
		api := &fakeMetaSchemaAPI{schemas: seededDefaults()}
		spec := []byte("- {name: organization, schema: '{\"type\":\"object\",\"required\":[\"cost_center\"]}'}\n")

		rep, err := NewMetaSchemaReconciler(api, "").Reconcile(context.Background(), spec, true)

		assert.NoError(t, err)
		assert.Equal(t, []string{"set metaschema organization"}, rep.Planned)
		assert.Equal(t, 0, rep.Applied)
		assert.Empty(t, api.updated)
	})

	t.Run("resets a built-in that holds an override and is left out", func(t *testing.T) {
		seeded := seededDefaults()
		for _, ms := range seeded {
			if ms.GetName() == metaschema.NameOrg {
				ms.Schema = `{"type":"string"}` // an override
			}
		}
		api := &fakeMetaSchemaAPI{schemas: seeded}

		rep, err := NewMetaSchemaReconciler(api, "").Reconcile(context.Background(), []byte("[]\n"), false)

		assert.NoError(t, err)
		assert.Equal(t, []string{"reset metaschema organization to default"}, rep.Planned)
		assert.Len(t, api.updated, 1)
		assert.Equal(t, metaschema.Defaults[metaschema.NameOrg], api.updated[0].schema)
	})

	t.Run("creates a built-in missing on the server", func(t *testing.T) {
		var seeded []*frontierv1beta1.MetaSchema
		for _, ms := range seededDefaults() {
			if ms.GetName() == metaschema.NameProspect {
				continue // server does not have it yet
			}
			seeded = append(seeded, ms)
		}
		api := &fakeMetaSchemaAPI{schemas: seeded}

		rep, err := NewMetaSchemaReconciler(api, "").Reconcile(context.Background(), []byte("[]\n"), false)

		assert.NoError(t, err)
		assert.Equal(t, []string{"create metaschema prospect"}, rep.Planned)
		assert.Len(t, api.created, 1)
		assert.Equal(t, metaschema.NameProspect, api.created[0].GetName())
	})

	t.Run("rejects an unknown name at validate", func(t *testing.T) {
		err := NewMetaSchemaReconciler(&fakeMetaSchemaAPI{}, "").Validate([]byte("- {name: widget, schema: '{}'}\n"))
		assert.ErrorContains(t, err, "unknown metaschema")
	})

	// R2 value model: there is no delete, so a delete flag is an unknown field.
	t.Run("R2 rejects a delete flag since metaschemas are values", func(t *testing.T) {
		err := NewMetaSchemaReconciler(&fakeMetaSchemaAPI{}, "").Validate([]byte("- {name: organization, schema: '{\"type\":\"object\"}', delete: true}\n"))
		assert.Error(t, err)
	})

	// R3 check the whole file first: one bad entry fails validation for the doc,
	// so nothing applies.
	t.Run("R3 a bad entry fails validation for the whole document", func(t *testing.T) {
		spec := []byte("- {name: organization, schema: '{\"type\":\"object\"}'}\n- {name: widget, schema: '{\"type\":\"object\"}'}\n")
		err := NewMetaSchemaReconciler(&fakeMetaSchemaAPI{}, "").Validate(spec)
		assert.ErrorContains(t, err, "unknown metaschema")
	})

	t.Run("exports only overridden built-ins", func(t *testing.T) {
		seeded := seededDefaults()
		for _, ms := range seeded {
			if ms.GetName() == metaschema.NameOrg {
				ms.Schema = `{"type":"object","required":["cost_center"]}`
			}
		}
		api := &fakeMetaSchemaAPI{schemas: seeded}

		out, err := NewMetaSchemaReconciler(api, "").Export(context.Background())

		assert.NoError(t, err)
		specs, ok := out.([]MetaSchemaSpec)
		assert.True(t, ok)
		assert.Equal(t, []MetaSchemaSpec{{Name: metaschema.NameOrg, Schema: `{"type":"object","required":["cost_center"]}`}}, specs)
	})
}
