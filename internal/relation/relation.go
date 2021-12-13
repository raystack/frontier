package relation

import "errors"

var (
	RelationDoesntExist = errors.New("relation doesn't exist")
	InvalidUUID         = errors.New("invalid syntax of uuid")
)
