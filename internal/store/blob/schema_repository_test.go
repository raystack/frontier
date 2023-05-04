package blob

import (
	"context"
	"testing"

	"github.com/odpf/shield/internal/bootstrap/schema"

	"github.com/stretchr/testify/assert"
)

func TestGetSchema(t *testing.T) {
	testBucket, err := NewStore(context.Background(), "file://testdata", "")
	assert.NoError(t, err)

	s := SchemaRepository{
		bucket: testBucket,
	}

	config, err := s.GetSchemas(context.Background())
	assert.NoError(t, err)

	expectedMap := []schema.ServiceDefinition{
		{
			Name: "potato",
			Resources: []schema.DefinitionResource{
				{
					Name: "order",
					Permissions: []schema.ResourcePermission{
						{
							Name: "delete",
						},
						{
							Name: "update",
						},
						{
							Name: "get",
						},
						{
							Name: "list",
						},
						{
							Name: "create",
						},
					},
				},
			},
		},
	}
	assert.Equal(t, expectedMap, config)
}
