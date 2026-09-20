package resource

type Filter struct {
	ProjectID     string
	UserID        string
	ServiceUserID string
	NamespaceID   string
	// IncludeDeleted also returns rows that have deleted_at set.
	// TODO(fix): remove once the project delete cascade no longer purges resources
	IncludeDeleted bool
}
