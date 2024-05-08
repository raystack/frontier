package crypt

import (
	"crypto/rand"
	"math/big"
)

// GenerateRandomStringFromLetters generates a random string of the given length using the provided runes
// this function panics if
// - the provided length is less than 1
// - if the provided runes are empty
// - if os fails to read random bytes
func GenerateRandomStringFromLetters(length int, letterRunes []rune) string {
	b := make([]rune, length)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letterRunes))))
		if err != nil {
			panic(err)
		}
		b[i] = letterRunes[num.Int64()]
	}
	return string(b)
}
