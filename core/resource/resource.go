package resource

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/odpf/shield/core/group"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/project"
	"github.com/odpf/shield/core/user"
)

const NON_RESOURCE_ID = "*"

var (
	ErrNotExist    = errors.New("resource doesn't exist")
	ErrInvalidUUID = errors.New("invalid syntax of uuid")
)

type Store interface {
	GetResource(ctx context.Context, id string) (Resource, error)
	GetResourceByURN(ctx context.Context, urn string) (Resource, error)
	CreateResource(ctx context.Context, resource Resource) (Resource, error)
	ListResources(ctx context.Context, filters Filters) ([]Resource, error)
	UpdateResource(ctx context.Context, id string, resource Resource) (Resource, error)
}

type AuthzStore interface {
	DeleteSubjectRelations(ctx context.Context, resourceType, optionalResourceID string) error
}

type BlobStore interface {
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

type Filters struct {
	ProjectID      string `json:"project_id"`
	GroupID        string `json:"group_id"`
	OrganizationID string `json:"org_id"`
	NamespaceID    string `json:"namespace_id"`
}

type YAML struct {
	Name         string              `json:"name" yaml:"name"`
	Backend      string              `json:"backend" yaml:"backend"`
	ResourceType string              `json:"resource_type" yaml:"resource_type"`
	Actions      map[string][]string `json:"actions" yaml:"actions"`
}

/*
 /project/uuid/
*/
func CreateURN(res Resource) string {
	isSystemNS := namespace.IsSystemNamespaceID(res.NamespaceID)
	if isSystemNS {
		return res.Name
	}
	if res.Name == NON_RESOURCE_ID {
		return fmt.Sprintf("p/%s/%s", res.ProjectID, res.NamespaceID)
	}
	return fmt.Sprintf("r/%s/%s", res.NamespaceID, res.Name)
}
