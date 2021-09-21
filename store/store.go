package store

import (
	"context"
	"errors"

	"github.com/odpf/shield/structs"
	"gocloud.dev/blob"
)

var (
	ErrResourceNotFound = errors.New("resource not found")
)

type RuleRepository interface {
	GetAll(ctx context.Context) ([]structs.Ruleset, error)
}

type Bucket interface {
	WriteAll(ctx context.Context, key string, p []byte, opts *blob.WriterOptions) error
	ReadAll(ctx context.Context, key string) ([]byte, error)
	List(opts *blob.ListOptions) *blob.ListIterator
	Delete(ctx context.Context, key string) error
	Close() error
}
