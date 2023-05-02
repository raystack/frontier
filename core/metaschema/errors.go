package metaschema

import "errors"

var (
	ErrInvalidName       = errors.New("metadata name is invalid")
	ErrNotExist          = errors.New("metadata doesn't exist")
	ErrConflict          = errors.New("metadata already exist")
	ErrInvalidDetail     = errors.New("invalid metadata detail")
	ErrInvalidMetaSchema = errors.New("metadata schema validation failed")
)
