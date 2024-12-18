package project

import "github.com/raystack/frontier/pkg/pagination"

type Filter struct {
	OrgID           string
	WithMemberCount bool
	ProjectIDs      []string
	State           State
	// NonInherited filters out projects that are inherited from access given through an organization
	NonInherited bool
	Pagination   *pagination.Pagination
}
