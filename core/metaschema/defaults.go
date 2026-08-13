package metaschema

import _ "embed"

// Built-in metaschema names. These are the schemas the server validates entity
// metadata against, and the set the MetaSchema reconcile kind manages.
const (
	NameUser     = "user"
	NameGroup    = "group"
	NameOrg      = "organization"
	NameRole     = "role"
	NameProspect = "prospect"
)

//go:embed metaschemas/user.json
var defaultUser []byte

//go:embed metaschemas/group.json
var defaultGroup []byte

//go:embed metaschemas/org.json
var defaultOrg []byte

//go:embed metaschemas/role.json
var defaultRole []byte

//go:embed metaschemas/prospect.json
var defaultProspect []byte

// Defaults maps each built-in metaschema name to its shipped JSON schema. It is
// the one source for both the server seeding (MigrateDefaults) and the MetaSchema
// reconcile kind, so adding a metaschema for a new resource is a single edit here.
var Defaults = map[string]string{
	NameUser:     string(defaultUser),
	NameGroup:    string(defaultGroup),
	NameOrg:      string(defaultOrg),
	NameRole:     string(defaultRole),
	NameProspect: string(defaultProspect),
}
