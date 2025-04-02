package store

import (
	"context"
	"errors"
	"github.com/odpf/shield/structs"
	"io"
)

var (
	ErrResourceNotFound = errors.New("resource not found")
)

type RuleRepository interface {
	GetAll(ctx context.Context) ([]structs.Ruleset, error)
}

type Bucket interface {
	ListFiles(ctx context.Context, prefix string) ([]string, error)
	ReadFile(ctx context.Context, key string) ([]byte, error)
	WriteFile(ctx context.Context, key string, data io.Reader) error
	Close() error
}
