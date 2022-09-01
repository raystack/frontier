package resource

import (
	"context"
	"fmt"
	"time"

	"github.com/odpf/shield/core/group"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/project"
	"github.com/odpf/shield/core/user"
)

const NON_RESOURCE_ID = "*"

type Repository interface {
	GetByID(ctx context.Context, id string) (Resource, error)
	GetByURN(ctx context.Context, urn string) (Resource, error)
	Create(ctx context.Context, resource Resource) (Resource, error)
	List(ctx context.Context, flt Filter) ([]Resource, error)
	Update(ctx context.Context, id string, resource Resource) (Resource, error)
}

type ConfigRepository interface {
	GetAll(ctx context.Context) ([]YAML, error)
	GetRelationsForNamespace(ctx context.Context, namespaceID string) (map[string]bool, error)
}

type Resource struct {
	Idxa           string
	URN            string
	Name           string
	ProjectID      string `json:"project_id"`
	Project        project.Project
	GroupID        string `json:"group_id"`
	Group          group.Group
	OrganizationID string `json:"organization_id"`
	Organization   organization.Organization
	NamespaceID    string `json:"namespace_id"`
	Namespace      namespace.Namespace
	User           user.User
	UserID         string `json:"user_id"`
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

type Filter struct {
	ProjectID      string
	GroupID        string
	OrganizationID string
	NamespaceID    string
}

type YAML struct {
	Name    string              `json:"name" yaml:"name"`
	Actions map[string][]string `json:"actions" yaml:"actions"`
}
