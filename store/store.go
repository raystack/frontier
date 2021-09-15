package store

import (
	"errors"
	"github.com/odpf/shield/structs"
)

var (
	ErrResourceNotFound = errors.New("resource not found")
)

type RuleRepository interface {
	GetAll() ([]structs.Service, error)
}
