package bootstrap_test

import (
	"context"
	_ "embed"
	"fmt"
	"sort"
	"testing"

	"github.com/authzed/spicedb/pkg/schemadsl/compiler"
	"github.com/odpf/shield/internal/bootstrap"
	"github.com/odpf/shield/internal/bootstrap/schema"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/compute_service.yml
var atlasSchemaYaml []byte

//go:embed testdata/compiled_schema.zed
var compiledSchemaZed string

func TestCompileSchema(t *testing.T) {
	tenantName := "acme"
	compiledSchema, err := compiler.Compile(compiler.InputSchema{
		Source:       "base_schema.zed",
		SchemaString: schema.BaseSchemaZed,
	}, &tenantName)
	assert.NoError(t, err)

	appService, err := bootstrap.BuildServiceDefinitionFromAZSchema(compiledSchema.ObjectDefinitions, "app")
	assert.NoError(t, err)
	assert.Len(t, appService.Permissions, 20)
}

func TestAddServiceToSchema(t *testing.T) {
	tenantName := "shield"
	existingSchema, err := compiler.Compile(compiler.InputSchema{
		Source:       "base_schema.zed",
		SchemaString: schema.BaseSchemaZed,
	}, &tenantName)
	assert.NoError(t, err)

	atlasServiceDefinition := schema.ServiceDefinition{}
	err = yaml.Unmarshal(atlasSchemaYaml, &atlasServiceDefinition)
	assert.NoError(t, err)

	spiceDBDefinitions := existingSchema.ObjectDefinitions
	spiceDBDefinitions, err = bootstrap.ApplyServiceDefinitionOverAZSchema(&atlasServiceDefinition, spiceDBDefinitions)
	assert.NoError(t, err)

	// sort definitions, useful to keep it consistent
	for idx := range spiceDBDefinitions {
		sort.Slice(spiceDBDefinitions[idx].Relation, func(i, j int) bool {
			return spiceDBDefinitions[idx].Relation[i].Name < spiceDBDefinitions[idx].Relation[j].Name
		})
	}
	sort.Slice(spiceDBDefinitions, func(i, j int) bool {
		return spiceDBDefinitions[i].Name < spiceDBDefinitions[j].Name
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
