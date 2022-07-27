package namespace

import (
	"context"
	"fmt"
	"time"
)

type Repository interface {
	Get(ctx context.Context, id string) (Namespace, error)
	Create(ctx context.Context, ns Namespace) (Namespace, error)
	List(ctx context.Context) ([]Namespace, error)
	Update(ctx context.Context, ns Namespace) (Namespace, error)
}

type Namespace struct {
	ID        string
	Name      string
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

//postgres://shield:@:5432/
func CreateID(backend, resourceType string) string {
	return fmt.Sprintf("%s_%s", backend, resourceType)
}
