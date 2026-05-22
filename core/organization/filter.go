package organization

import (
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/pkg/pagination"
)

type Filter struct {
	// Principal restricts results to orgs the principal has a policy on.
	// Intersected with IDs when both are set.
	Principal *authenticate.Principal

	IDs   []string
	State State

	Pagination *pagination.Pagination
}
