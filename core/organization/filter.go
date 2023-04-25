package organization

type Filter struct {
	// only one filter gets applied at a time

	UserID string
	State  State
}
