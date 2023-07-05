package serviceuser

type Filter struct {
	ServiceUserID string
	OrgID         string
	IsKey         bool
	IsSecret      bool
	State         State
}
