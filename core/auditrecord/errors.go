package auditrecord

import "errors"

var (
	ErrIdempotencyKeyConflict = errors.New("audit record already exists for the given idempotency key")
	ErrInvalidUUID            = errors.New("invalid syntax of uuid")
	ErrNotFound               = errors.New("audit record not found")
	ErrRepositoryBadInput     = errors.New("invalid repository input")
)
