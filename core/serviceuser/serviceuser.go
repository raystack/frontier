package serviceuser

import (
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/raystack/frontier/pkg/metadata"
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
	ID       string
	OrgID    string
	Title    string
	State    string
	Metadata metadata.Metadata

	// CreatedByUser is a transient field that is used to track the user who created this service user
	// this doesn't have any impact on the service user itself
	CreatedByUser string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type CredentialType string

func (c CredentialType) String() string {
	return string(c)
}

const (
	ClientSecretCredentialType CredentialType = "client_credential"
	JWTCredentialType          CredentialType = "jwt_bearer"
	OpaqueTokenCredentialType  CredentialType = "opaque_token"
)

type Credential struct {
	// ID is the unique identifier of the credential.
	// This is also used as kid in JWT, the spec doesn't
	// state how the kid should be generated as anyway this token
	// is owned by frontier, and we are in control of key generation
	// any arbitrary string can be used as kid as long as its unique
	ID            string
	ServiceUserID string
	Type          CredentialType

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

type Secret struct {
	ID        string
	Title     string
	Value     string
	CreatedAt time.Time
}

type Token struct {
	ID        string
	Title     string
	Value     string
	CreatedAt time.Time
}
