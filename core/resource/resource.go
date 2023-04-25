package resource

import (
	"context"
	"fmt"
	"time"

	"github.com/odpf/shield/core/namespace"
)

const NON_RESOURCE_ID = "*"

type Repository interface {
	GetByID(ctx context.Context, id string) (Resource, error)
	GetByURN(ctx context.Context, urn string) (Resource, error)
	Create(ctx context.Context, resource Resource) (Resource, error)
	List(ctx context.Context, flt Filter) ([]Resource, error)
	Update(ctx context.Context, id string, resource Resource) (Resource, error)
	GetByNamespace(ctx context.Context, name string, ns string) (Resource, error)
	Delete(ctx context.Context, id string) error
}

type ConfigRepository interface {
	GetAll(ctx context.Context) ([]YAML, error)
}

type Resource struct {
	ID             string
	URN            string
	Name           string
	ProjectID      string
	OrganizationID string
	NamespaceID    string
	UserID         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (res Resource) CreateURN() string {
	isSystemNS := namespace.IsSystemNamespaceID(res.NamespaceID)
	if isSystemNS {
		return res.Name
	}
	if res.Name == NON_RESOURCE_ID {
		return fmt.Sprintf("p/%s/%s", res.ProjectID, res.NamespaceID)
	}
	return fmt.Sprintf("r/%s/%s", res.NamespaceID, res.Name)
}

type YAML struct {
	Name         string              `json:"name" yaml:"name"`
	Backend      string              `json:"backend" yaml:"backend"`
	ResourceType string              `json:"resource_type" yaml:"resource_type"`
	Actions      map[string][]string `json:"actions" yaml:"actions"`
}
