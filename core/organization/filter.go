package organization

import (
	"github.com/raystack/frontier/pkg/pagination"
)

type Filter struct {
	UserID string

	IDs   []string
	State State

	Pagination *pagination.Pagination
}
