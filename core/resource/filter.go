package resource

type Filter struct {
	ProjectID     string
	UserID        string
	ServiceUserID string
	NamespaceID   string
	// IncludeDeleted also returns soft-deleted rows.
	// TODO(fix): remove once the project delete cascade no longer purges
	IncludeDeleted bool
}
