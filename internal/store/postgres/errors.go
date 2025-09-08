package postgres

import "errors"

var (
	ErrDuplicateKey              = errors.New("duplicate key")
	ErrCheckViolation            = errors.New("check constraint violation")
	ErrForeignKeyViolation       = errors.New("foreign key violation")
	ErrInvalidTextRepresentation = errors.New("invalid input syntax type")
	ErrInvalidID                 = errors.New("invalid id")
	ErrImmutableRecord           = errors.New("record is immutable and cannot be modified")
)
