package uuid

import "github.com/google/uuid"

func IsValid(key string) bool {
	_, err := uuid.Parse(key)
	return err == nil
}
