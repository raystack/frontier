package policy

type Filter struct {
	OrgID     string
	ProjectID string
	GroupID   string
	RoleID    string
	RoleIDs   []string

	PrincipalType string
	PrincipalID   string
	PrincipalIDs  []string
	ResourceType  string
}
