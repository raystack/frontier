package consent

import "errors"

var (
	// ErrUnknownDocuments is returned for an id this deployment does not
	// configure, which is how a mismatch with the client's list surfaces.
	ErrUnknownDocuments = errors.New("unknown consent document ids")

	// ErrMissingDocuments is returned when the ids do not cover every configured
	// document. All of them are required at signup.
	ErrMissingDocuments = errors.New("missing consent document ids")

	// ErrInvalidGrant is returned when a grant is missing something the record needs.
	ErrInvalidGrant = errors.New("invalid consent grant")

	// ErrConsentExists is returned when a user already has a signup record.
	ErrConsentExists = errors.New("a consent record already exists for this user")
)
