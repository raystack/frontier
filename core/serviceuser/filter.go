package serviceuser

type Filter struct {
	ServiceUserID  string
	ServiceUserIDs []string

	OrgID    string
	IsKey    bool
	IsSecret bool
	State    State
}
