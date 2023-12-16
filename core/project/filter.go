package project

type Filter struct {
	OrgID           string
	NonInherited    bool
	WithMemberCount bool
	ProjectIDs      []string
	State           State
}
