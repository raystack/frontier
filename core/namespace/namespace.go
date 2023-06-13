package namespace

import (
	"context"
	"fmt"
	"time"

	"github.com/raystack/shield/pkg/metadata"
)

type Repository interface {
	Get(ctx context.Context, id string) (Namespace, error)
	Upsert(ctx context.Context, ns Namespace) (Namespace, error)
	List(ctx context.Context) ([]Namespace, error)
	Update(ctx context.Context, ns Namespace) (Namespace, error)
}

type Namespace struct {
	ID        string
	Name      string
	Metadata  metadata.Metadata
	CreatedAt time.Time
	UpdatedAt time.Time
}

func strListHas(list []string, a string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func IsSystemNamespaceID(nsID string) bool {
	return strListHas(systemIdsDefinition, nsID)
}

func CreateID(backend, resourceType string) string {
	if resourceType == "" {
		return backend
	}

	return fmt.Sprintf("%s/%s", backend, resourceType)
}
