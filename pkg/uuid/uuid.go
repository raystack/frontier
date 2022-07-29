package uuid

import "github.com/google/uuid"

// type alias
var NewString = uuid.NewString

func IsValid(key string) bool {
	_, err := uuid.Parse(key)
	return err == nil
}
