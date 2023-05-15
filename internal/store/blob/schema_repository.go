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

const (
	ServiceSuffix = ".service"
	RolesSuffix   = ".roles"
)

type SchemaRepository struct {
	bucket Bucket
}

func NewSchemaConfigRepository(b Bucket) *SchemaRepository {
	return &SchemaRepository{bucket: b}
}

func (s *SchemaRepository) GetRoles(ctx context.Context) ([]schema.DefinitionRoles, error) {
	var definitionRoles []schema.DefinitionRoles

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
		if !(strings.HasSuffix(obj.Key, fmt.Sprintf("%s.yaml", RolesSuffix)) || strings.HasSuffix(obj.Key, fmt.Sprintf("%s.yml", RolesSuffix))) {
			continue
		}
		fileBytes, err := s.bucket.ReadAll(ctx, obj.Key)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", "error in reading bucket object", err.Error())
		}

		var rf schema.RoleFile
		if err := yaml.Unmarshal(fileBytes, &rf); err != nil {
			return nil, errors.Wrap(err, "readServiceFiles: yaml.Unmarshal: "+obj.Key)
		}
		definitionRoles = append(definitionRoles, rf.Roles...)
	}

	return definitionRoles, nil
}

func (s *SchemaRepository) GetSchemas(ctx context.Context) ([]schema.ServiceDefinition, error) {
	var definitionMap []schema.ServiceDefinition

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
		if !(strings.HasSuffix(obj.Key, fmt.Sprintf("%s.yaml", ServiceSuffix)) || strings.HasSuffix(obj.Key, fmt.Sprintf("%s.yml", ServiceSuffix))) {
			continue
		}
		fileBytes, err := s.bucket.ReadAll(ctx, obj.Key)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", "error in reading bucket object", err.Error())
		}

		var def schema.ServiceDefinition
		if err := yaml.Unmarshal(fileBytes, &def); err != nil {
			return nil, errors.Wrap(err, "readServiceFiles: yaml.Unmarshal: "+obj.Key)
		}
		definitionMap = append(definitionMap, def)
	}

	return definitionMap, nil
}
