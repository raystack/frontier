package bootstrap_test

import (
	"context"
	_ "embed"
	"fmt"
	"testing"

	"github.com/authzed/spicedb/pkg/schemadsl/compiler"
	"github.com/odpf/shield/internal/bootstrap"
	"github.com/odpf/shield/internal/bootstrap/schema"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/potato.service.yml
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

	appService, err := bootstrap.BuildServiceDefinitionFromAZSchema("app", compiledSchema.ObjectDefinitions)
	assert.NoError(t, err)
	assert.Len(t, appService.Resources, 8)
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
	spiceDBDefinitions, err = bootstrap.ApplyServiceDefinitionOverAZSchema(atlasServiceDefinition, spiceDBDefinitions)
	assert.NoError(t, err)

	authzedSchemaSource, err := bootstrap.PrepareSchemaAsSource(spiceDBDefinitions)
	assert.NoError(t, err)

	// compile and validate generated schema
	err = bootstrap.ValidatePreparedAZSchema(context.Background(), authzedSchemaSource)
	assert.NoError(t, err)

	if compiledSchemaZed != authzedSchemaSource {
		fmt.Println(authzedSchemaSource)
	}
	assert.Equal(t, compiledSchemaZed, authzedSchemaSource)
}
