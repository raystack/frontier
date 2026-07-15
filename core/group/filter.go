package group

import "github.com/raystack/frontier/core/authenticate"

type Filter struct {
	// only one filter gets applied at a time

	SU bool // super user

	// Principal restricts results to groups the principal has a policy on.
	// Intersected with GroupIDs when both are set.
	Principal *authenticate.Principal

	OrganizationID  string
	State           State
	WithMemberCount bool

	IncludeDisabled bool

	GroupIDs []string
}
