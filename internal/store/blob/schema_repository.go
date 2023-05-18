package blob

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/odpf/shield/internal/bootstrap/schema"

	"github.com/pkg/errors"
	"gocloud.dev/blob"
	"gopkg.in/yaml.v3"
)

type SchemaRepository struct {
	bucket Bucket
}

func NewSchemaConfigRepository(b Bucket) *SchemaRepository {
	return &SchemaRepository{bucket: b}
}

// GetDefinition returns the service definition from the bucket
func (s *SchemaRepository) GetDefinition(ctx context.Context) (*schema.ServiceDefinition, error) {
	var definitions []schema.ServiceDefinition

	// iterate over bucket files, only read .yml & .yaml files
	it := s.bucket.List(&blob.ListOptions{})
	for {
		obj, err := it.Next(ctx)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if obj.IsDir {
			continue
		}
		if !(strings.HasSuffix(obj.Key, ".yaml") || strings.HasSuffix(obj.Key, ".yml")) {
			continue
		}
		fileBytes, err := s.bucket.ReadAll(ctx, obj.Key)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", "error in reading bucket object", err.Error())
		}

		var def schema.ServiceDefinition
		if err := yaml.Unmarshal(fileBytes, &def); err != nil {
			return nil, errors.Wrap(err, "GetDefinitions: yaml.Unmarshal: "+obj.Key)
		}
		definitions = append(definitions, def)
	}

	return schema.MergeServiceDefinitions(definitions...), nil
}
