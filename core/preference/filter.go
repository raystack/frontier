package preference

type Filter struct {
	OrgID        string `json:"org_id"`
	ProjectID    string `json:"project_id"`
	UserID       string `json:"user_id"`
	GroupID      string `json:"group_id"`
	ResourceID   string `json:"resource_id"`
	ResourceType string `json:"resource_type"`
}
