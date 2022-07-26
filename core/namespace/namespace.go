package namespace

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrNotExist    = errors.New("actions doesn't exist")
	ErrInvalidUUID = errors.New("invalid syntax of uuid")
)

type Store interface {
	GetNamespace(ctx context.Context, id string) (Namespace, error)
	CreateNamespace(ctx context.Context, namespace Namespace) (Namespace, error)
	ListNamespaces(ctx context.Context) ([]Namespace, error)
	UpdateNamespace(ctx context.Context, id string, namespace Namespace) (Namespace, error)
}

type Namespace struct {
	Id        string
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
