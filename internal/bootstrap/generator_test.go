package bootstrap_test

import (
	"context"
	_ "embed"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/authzed/spicedb/pkg/schemadsl/compiler"
	"github.com/raystack/frontier/internal/bootstrap"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/compute_service.yml
var computeSchemaYaml []byte

//go:embed testdata/compiled_schema.zed
var compiledSchemaZed string

func TestCompileSchema(t *testing.T) {
	tenantName := "acme"
	compiledSchema, err := compiler.Compile(compiler.InputSchema{
		Source:       "base_schema.zed",
		SchemaString: schema.BaseSchemaZed,
	}, compiler.ObjectTypePrefix(tenantName))
	assert.NoError(t, err)

	appService, err := bootstrap.BuildServiceDefinitionFromAZSchema(compiledSchema.ObjectDefinitions, "app")
	assert.NoError(t, err)
	assert.Len(t, appService.Permissions, 25)
}

func TestAddServiceToSchema(t *testing.T) {
	tenantName := "frontier"
	existingSchema, err := compiler.Compile(compiler.InputSchema{
		Source:       "base_schema.zed",
		SchemaString: schema.BaseSchemaZed,
	}, compiler.ObjectTypePrefix(tenantName))
	assert.NoError(t, err)

	computeServiceDefinition := schema.ServiceDefinition{}
	err = yaml.Unmarshal(computeSchemaYaml, &computeServiceDefinition)
	assert.NoError(t, err)

	spiceDBDefinitions := existingSchema.ObjectDefinitions
	spiceDBDefinitions, err = bootstrap.ApplyServiceDefinitionOverAZSchema(&computeServiceDefinition, spiceDBDefinitions)
	assert.NoError(t, err)

	// sort definitions, useful to keep it consistent
	for idx := range spiceDBDefinitions {
		sort.Slice(spiceDBDefinitions[idx].GetRelation(), func(i, j int) bool {
			return spiceDBDefinitions[idx].GetRelation()[i].GetName() < spiceDBDefinitions[idx].GetRelation()[j].GetName()
		})
	}
	sort.Slice(spiceDBDefinitions, func(i, j int) bool {
		return spiceDBDefinitions[i].GetName() < spiceDBDefinitions[j].GetName()
	})

	authzedSchemaSource, err := bootstrap.PrepareSchemaAsAZSource(spiceDBDefinitions)
	assert.NoError(t, err)

	// compile and validate generated schema
	err = bootstrap.ValidatePreparedAZSchema(context.Background(), authzedSchemaSource)
	assert.NoError(t, err)

	if compiledSchemaZed != authzedSchemaSource {
		fmt.Println(authzedSchemaSource)
	}
	assert.Equal(t, compiledSchemaZed, authzedSchemaSource)
}

func TestLongestPermissionSlugCompiles(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		want      bool
	}{
		{"sixty characters", strings.Repeat("a", 28) + "/" + strings.Repeat("b", 27), true},
		{"sixty-one characters", strings.Repeat("a", 28) + "/" + strings.Repeat("b", 28), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, schema.PermissionSlugWithinLimit(tt.namespace, "get"))

			defs, err := bootstrap.ApplyServiceDefinitionOverAZSchema(&schema.ServiceDefinition{
				Permissions: []schema.ResourcePermission{{Namespace: tt.namespace, Name: "get"}},
			}, bootstrap.GetBaseAZSchema())
			assert.NoError(t, err)
			source, err := bootstrap.PrepareSchemaAsAZSource(defs)
			assert.NoError(t, err)

			err = bootstrap.ValidatePreparedAZSchema(context.Background(), source)
			assert.Equal(t, tt.want, err == nil, "compile error: %v", err)
		})
	}
}
