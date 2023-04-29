package project

type Filter struct {
	// only one filter gets applied at a time

	OrgID string
	State State
}
