package policy

type Filter struct {
	OrgID     string
	ProjectID string
	GroupID   string
	RoleID    string

	PrincipalType string
	PrincipalID   string
	PrincipalIDs  []string
	ResourceType  string
}
