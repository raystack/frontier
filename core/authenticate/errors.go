package authenticate

import "errors"

var (
	ErrInvalidID = errors.New("user id is invalid")

	// errSkip signals that this authenticator doesn't apply to the request.
	errSkip = errors.New("skip authenticator")
)
