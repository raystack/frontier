package preference

type Filter struct {
	OrgID        string `json:"org_id"`
	ProjectID    string `json:"project_id"`
	UserID       string `json:"user_id"`
	GroupID      string `json:"group_id"`
	ResourceID   string `json:"resource_id"`
	ResourceType string `json:"resource_type"`
	// ScopeType and ScopeID filter preferences by their scope
	// e.g., to get user preferences scoped to a specific organization
	ScopeType string `json:"scope_type"`
	ScopeID   string `json:"scope_id"`
}
