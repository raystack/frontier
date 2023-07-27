package permission

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/raystack/frontier/pkg/metadata"
)

type Repository interface {
	Get(ctx context.Context, id string) (Permission, error)
	GetBySlug(ctx context.Context, id string) (Permission, error)
	Upsert(ctx context.Context, action Permission) (Permission, error)
	List(ctx context.Context, flt Filter) ([]Permission, error)
	Update(ctx context.Context, action Permission) (Permission, error)
	Delete(ctx context.Context, id string) error
}

type Permission struct {
	ID          string
	Name        string
	Slug        string
	NamespaceID string
	Metadata    metadata.Metadata

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (p Permission) GenerateSlug() string {
	return schema.FQPermissionNameFromNamespace(p.NamespaceID, p.Name)
}

func ParsePermissionName(s string) string {
	return convertDotPermissionToSlug(convertHashPermissionToSlug(convertColonPermissionToSlug(s)))
}

// convertHashPermissionToSlug will rarely be used but still if someone specifies a permission
// in the form of app/project#view it should process it correctly
func convertHashPermissionToSlug(s string) string {
	if strings.Contains(s, "/") {
		parts := strings.Split(s, "#")
		if len(parts) > 1 {
			subparts := strings.Split(parts[0], "/")
			return fmt.Sprintf("%s_%s_%s", subparts[0], subparts[1], parts[1])
		}
	}
	return s
}

// convertColonPermissionToSlug will rarely be used but still if someone specifies a permission
// in the form of app/project:view it should process it correctly
func convertColonPermissionToSlug(s string) string {
	if strings.Contains(s, "/") {
		parts := strings.Split(s, ":")
		if len(parts) > 1 {
			subparts := strings.Split(parts[0], "/")
			return fmt.Sprintf("%s_%s_%s", subparts[0], subparts[1], parts[1])
		}
	}
	return s
}

// convertDotPermissionToSlug if someone specifies a permission
// in the form of app.project.view it should process it correctly
func convertDotPermissionToSlug(s string) string {
	if strings.Contains(s, ".") {
		parts := strings.Split(s, ".")
		if len(parts) == 3 {
			return fmt.Sprintf("%s_%s_%s", parts[0], parts[1], parts[2])
		}
	}
	return s
}
