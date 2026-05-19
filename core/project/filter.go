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

	// OrgIDs narrows results to projects whose org_id is in this list. Used by
	// membership listing to batch-expand all projects across the orgs a
	// principal can inherit project visibility from. If both OrgID and OrgIDs
	// are set, projects must satisfy both (intersection) — typically yields
	// no rows unless OrgID is one of OrgIDs.
	OrgIDs []string
}
