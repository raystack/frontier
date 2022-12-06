package action

import (
	"context"
	"time"
)

type Repository interface {
	Get(ctx context.Context, id string) (Action, error)
	Create(ctx context.Context, action Action) (Action, error)
	List(ctx context.Context) ([]Action, error)
	Update(ctx context.Context, action Action) (Action, error)
}

type Action struct {
	ID          string
	Name        string
	NamespaceID string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
