package group

type Filter struct {
	// only one filter gets applied at a time

	SU bool // super user

	OrganizationID  string
	State           State
	WithMemberCount bool

	GroupIDs []string
}
