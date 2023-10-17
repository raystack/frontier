package organization

type Filter struct {
	UserID string

	IDs   []string
	State State
}
