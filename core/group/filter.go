package group

type Filter struct {
	// only one filter gets applied at a time

	OrganizationID  string
	State           State
	WithMemberCount bool

	GroupIDs []string
}
