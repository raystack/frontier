package utils

import (
	"net/mail"

	"github.com/google/uuid"
)

// NewString is type alias to `github.com/google/uuid`.NewString
var NewString = uuid.NewString

// IsValidUUID returns true if passed string in uuid format
// defined by `github.com/google/uuid`.Parse
// else return false
func IsValidUUID(key string) bool {
	_, err := uuid.Parse(key)
	return err == nil
}

// IsNullUUID returns true if passed string is a null uuid or is not a valid uuid
// defined by `github.com/google/uuid`.Parse and `github.com/google/uuid`.Nil respectively
// else return false
func IsNullUUID(key string) bool {
	k, err := uuid.Parse(key)
	return err != nil || k == uuid.Nil
}

func IsValidEmail(str string) bool {
	_, err := mail.ParseAddress(str)
	return err == nil
}
