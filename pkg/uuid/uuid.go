package uuid

import "github.com/google/uuid"

// NewString is type alias to `github.com/google/uuid`.NewString
var NewString = uuid.NewString

// IsValid returns true if passed string in uuid format
// defined by `github.com/google/uuid`.Parse
// else return false
func IsValid(key string) bool {
	_, err := uuid.Parse(key)
	return err == nil
}
