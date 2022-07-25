package role

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/odpf/shield/core/namespace"
)

var (
	ErrNotExist    = errors.New("role doesn't exist")
	ErrInvalidUUID = errors.New("invalid syntax of uuid")
)

type Repository interface {
	Create(ctx context.Context, role Role) (Role, error)
	Get(ctx context.Context, id string) (Role, error)
	List(ctx context.Context) ([]Role, error)
	Update(ctx context.Context, toUpdate Role) (Role, error)
}

type Role struct {
	ID          string
	Name        string
	Types       []string
	Namespace   namespace.Namespace
	NamespaceID string
	Metadata    map[string]any
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func GetOwnerRole(ns namespace.Namespace) Role {
	id := fmt.Sprintf("%s_%s", ns.ID, "owner")
	name := fmt.Sprintf("%s_%s", strings.Title(ns.ID), "Owner")
	return Role{
		ID:        id,
		Name:      name,
		Types:     []string{UserType},
		Namespace: ns,
	}
}
