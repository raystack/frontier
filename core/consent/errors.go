package consent

import "errors"

var (
	// ErrUnknownDocuments is returned for an id this deployment does not
	// configure, which is how a mismatch with the client's list surfaces.
	ErrUnknownDocuments = errors.New("unknown consent document ids")

	// ErrMissingDocuments is returned when the ids do not cover every configured
	// document. All of them are required at signup.
	ErrMissingDocuments = errors.New("missing consent document ids")
)
