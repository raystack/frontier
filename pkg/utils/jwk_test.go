package utils

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
)

func TestBuildToken(t *testing.T) {
	issuer := "test"
	sub := uuid.New().String()
	validity := time.Minute * 10
	kid := uuid.New().String()
	newKey, err := CreateJWKWithKID(kid)
	assert.NoError(t, err)
	t.Run("create a valid token", func(t *testing.T) {
		got, err := BuildToken(newKey, issuer, sub, validity, nil)
		assert.NoError(t, err)
		parsedToken, err := jwt.ParseInsecure(got)
		assert.NoError(t, err)
		assert.Equal(t, issuer, parsedToken.Issuer())
		assert.Equal(t, sub, parsedToken.Subject())
		gotKid, _ := parsedToken.Get(jwk.KeyIDKey)
		assert.Equal(t, kid, gotKid)
	})
}
