package metaschema

import "errors"

var (
	ErrInvalidID         = errors.New("metaschema id is invalid")
	ErrNotExist          = errors.New("metaschema doesn't exist")
	ErrConflict          = errors.New("metaschema already exist")
	ErrInvalidDetail     = errors.New("invalid metadata detail")
	ErrInvalidMetaSchema = errors.New("metadata schema validation failed")
)
