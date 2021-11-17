package schema

import (
	"context"
	"errors"
)

type Service struct {
	Store Store
}

var InvalidUUID = errors.New("invalid syntax of uuid")

type Store interface {
	GetAction(ctx context.Context, id string) (Action, error)
	CreateAction(ctx context.Context, action Action) (Action, error)
	ListActions(ctx context.Context) ([]Action, error)
}
