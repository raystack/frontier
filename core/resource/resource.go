package resource

import (
	"context"
	"fmt"
	"time"

	"github.com/raystack/frontier/core/relation"

	"github.com/raystack/frontier/pkg/metadata"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (Resource, error)
	GetByURN(ctx context.Context, urn string) (Resource, error)
	Create(ctx context.Context, resource Resource) (Resource, error)
	List(ctx context.Context, flt Filter) ([]Resource, error)
	Update(ctx context.Context, resource Resource) (Resource, error)
	Delete(ctx context.Context, id string) error
}

type ConfigRepository interface {
	GetAll(ctx context.Context) ([]YAML, error)
}

type Resource struct {
	ID            string `json:"id"`
	URN           string `json:"urn"`
	Name          string `json:"name"`
	Title         string `json:"title"`
	ProjectID     string `json:"project_id"`
	NamespaceID   string `json:"namespace_id"`
	PrincipalID   string `json:"principal_id"`
	PrincipalType string `json:"principal_type"`
	Metadata      metadata.Metadata

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (res Resource) CreateURN(projectName string) string {
	return fmt.Sprintf("frn:%s:%s:%s", projectName, res.NamespaceID, res.Name)
}

type YAML struct {
	Name         string              `json:"name" yaml:"name"`
	Backend      string              `json:"backend" yaml:"backend"`
	ResourceType string              `json:"resource_type" yaml:"resource_type"`
	Actions      map[string][]string `json:"actions" yaml:"actions"`
}

type Check struct {
	Object     relation.Object
	Subject    relation.Subject
	Permission string
}
