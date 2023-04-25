package user

type Filter struct {
	// only one filter gets applied at a time

	Limit int32
	Page  int32

	Keyword string
	OrgID   string
	GroupID string
	State   State
}
