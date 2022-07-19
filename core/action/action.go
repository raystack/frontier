package action

import (
	"context"
	"errors"
	"time"

	"github.com/odpf/shield/core/namespace"
)

var (
	ErrInvalidUUID = errors.New("invalid syntax of uuid")
	ErrNotExist    = errors.New("actions doesn't exist")
)

type Store interface {
	GetAction(ctx context.Context, id string) (Action, error)
	CreateAction(ctx context.Context, action Action) (Action, error)
	ListActions(ctx context.Context) ([]Action, error)
	UpdateAction(ctx context.Context, action Action) (Action, error)
}

type Action struct {
	Id          string
	Name        string
	NamespaceId string
	Namespace   namespace.Namespace
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
