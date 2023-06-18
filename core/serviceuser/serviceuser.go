package serviceuser

import (
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/raystack/shield/pkg/metadata"
)

const (
	DefaultKeyType = "sv_rsa"
)

type State string

func (s State) String() string {
	return string(s)
}

const (
	Enabled  State = "enabled"
	Disabled State = "disabled"
)

type ServiceUser struct {
	ID        string
	OrgID     string
	Title     string
	State     string
	Metadata  metadata.Metadata
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Credential struct {
	// ID is the unique identifier of the credential.
	// This is also used as kid in JWT, the spec doesn't
	// state how the kid should be generated as anyway this token
	// is owned by shield, and we are in control of key generation
	// any arbitrary string can be used as kid as long as its unique
	ID            string
	ServiceUserID string

	// SecretHash used for basic auth
	SecretHash string

	// PublicKey used for JWT verification using RSA
	PublicKey jwk.Set
	// PrivateKey used for JWT signing using RSA, this is not stored and
	// only generated and returned when creating a new credential
	PrivateKey []byte

	Title     string
	Metadata  metadata.Metadata
	CreatedAt time.Time
	UpdatedAt time.Time
}
