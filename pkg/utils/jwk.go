package utils

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/lestrrat-go/jwx/v2/jwk"
)

const (
	RSAKeySize = 2048
)

func CreateJWKs(numOfKeys int) (jwk.Set, error) {
	keySet := jwk.NewSet()
	for ; numOfKeys > 0; numOfKeys-- {
		// generate keys
		keyRaw, err := rsa.GenerateKey(rand.Reader, RSAKeySize)
		if err != nil {
			return nil, err
		}
		rsaKey, err := jwk.FromRaw(keyRaw)
		if err != nil {
			return nil, err
		}
		pubKey, err := rsaKey.PublicKey()
		if err != nil {
			return nil, err
		}
		thumb, err := pubKey.Thumbprint(crypto.SHA256)
		if err != nil {
			return nil, err
		}
		rsaKey.Set(jwk.AlgorithmKey, "RS256")
		rsaKey.Set(jwk.KeyUsageKey, "sig")
		rsaKey.Set(jwk.KeyIDKey, base64.RawURLEncoding.EncodeToString(thumb))
		keySet.AddKey(rsaKey)
	}
	return keySet, nil
}

func CreateJWKWithKID(id string) (jwk.Key, error) {
	// generate key
	keyRaw, err := rsa.GenerateKey(rand.Reader, RSAKeySize)
	if err != nil {
		return nil, err
	}
	rsaKey, err := jwk.FromRaw(keyRaw)
	if err != nil {
		return nil, err
	}
	rsaKey.Set(jwk.AlgorithmKey, "RS256")
	rsaKey.Set(jwk.KeyUsageKey, "sig")
	rsaKey.Set(jwk.KeyIDKey, id)
	return rsaKey, nil
}

// GetPublicKeySet convert private to public
func GetPublicKeySet(ctx context.Context, privateKeySet jwk.Set) (jwk.Set, error) {
	publicKeySet := jwk.NewSet()
	for iter := privateKeySet.Keys(ctx); iter.Next(ctx); {
		pair := iter.Pair()
		key := pair.Value.(jwk.Key)

		pubKey, err := key.PublicKey()
		if err != nil {
			return nil, fmt.Errorf("failed to generate public key from private rsa: %w", err)
		}
		publicKeySet.AddKey(pubKey)
	}
	return publicKeySet, nil
}

// BuildToken creates a signed jwt using provided private key
// Ensure the key contains kid else the operation fails
func BuildToken(rsaKey jwk.Key, issuer, sub string,
	validity time.Duration, customClaims map[string]string) ([]byte, error) {
	if rsaKey.KeyID() == "" {
		return nil, fmt.Errorf("key id is empty")
	}
	body := jwt.NewBuilder().
		Issuer(issuer).
		IssuedAt(time.Now().UTC()).
		NotBefore(time.Now().UTC()).
		Expiration(time.Now().UTC().Add(validity)).
		JwtID(uuid.New().String()).
		Subject(sub)
	body.Claim(jwk.KeyIDKey, rsaKey.KeyID())
	for claimKey, claimVal := range customClaims {
		body = body.Claim(claimKey, claimVal)
	}

	tok, err := body.Build()
	if err != nil {
		return nil, err
	}

	return jwt.Sign(tok, jwt.WithKey(jwa.RS256, rsaKey))
}
