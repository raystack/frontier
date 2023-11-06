package token

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/raystack/frontier/pkg/utils"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

var (
	ErrMissingRSADisableToken = errors.New("rsa key missing in config, generate and pass file path")
	ErrInvalidToken           = errors.New("failed to verify a valid token")
)

const (
	GeneratedClaimKey   = "gen"
	GeneratedClaimValue = "system"
	OrgIDsClaimKey      = "org_ids"
)

type Service struct {
	keySet       jwk.Set
	publicKeySet jwk.Set
	issuer       string
	validity     time.Duration
}

// NewService creates a new token service
// generate keys used for rsa via frontier cli "frontier server keygen"
func NewService(keySet jwk.Set, issuer string, validity time.Duration) Service {
	publicKeySet := jwk.NewSet()
	if keySet != nil {
		pub, err := utils.GetPublicKeySet(context.Background(), keySet)
		if err != nil {
			panic(err)
		}
		publicKeySet = pub
	}

	return Service{
		keySet:       keySet,
		issuer:       issuer,
		publicKeySet: publicKeySet,
		validity:     validity,
	}
}

// GetPublicKeySet returns the public keys to verify the access token
func (s Service) GetPublicKeySet() jwk.Set {
	return s.publicKeySet
}

// Build creates an access token for the given subjectID
func (s Service) Build(subjectID string, metadata map[string]string) ([]byte, error) {
	if s.keySet == nil {
		return nil, ErrMissingRSADisableToken
	}
	// use first key to sign token
	rsaKey, ok := s.keySet.Key(0)
	if !ok {
		return nil, errors.New("missing rsa key to generate token")
	}

	// frontier generated token has an extra custom claim
	// used to identify which public key to use to verify the token
	metadata[GeneratedClaimKey] = GeneratedClaimValue
	return utils.BuildToken(rsaKey, s.issuer, subjectID, s.validity, metadata)
}

func (s Service) Parse(ctx context.Context, userToken []byte) (string, map[string]any, error) {
	if s.keySet == nil {
		return "", nil, ErrMissingRSADisableToken
	}
	// verify token with jwks
	verifiedToken, err := jwt.Parse(userToken, jwt.WithKeySet(s.publicKeySet))
	if err != nil {
		return "", nil, fmt.Errorf("%s: %w", ErrInvalidToken.Error(), err)
	}
	tokenClaims, err := verifiedToken.AsMap(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("%s: %w", ErrInvalidToken.Error(), err)
	}
	return verifiedToken.Subject(), tokenClaims, nil
}
