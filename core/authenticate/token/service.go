package token

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/raystack/shield/pkg/utils"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

var (
	ErrMissingRSADisableToken = errors.New("rsa key missing in config, generate and pass file path")
	ErrInvalidToken           = errors.New("failed to verify a valid token")
	ErrNoToken                = errors.New("no token")
)

const (
	// TODO(kushsharma): should we expose this in config?
	tokenValidity = time.Hour * 24 * 7

	GeneratedClaimKey   = "gen"
	GeneratedClaimValue = "system"
)

type Service struct {
	keySet       jwk.Set
	publicKeySet jwk.Set
	issuer       string
}

// NewService creates a new token service
// generate keys used for rsa via shield cli "shield server keygen"
func NewService(keySet jwk.Set, issuer string) Service {
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
	}
}

func (s Service) GetPublicKeySet() jwk.Set {
	return s.publicKeySet
}

func (s Service) Build(userID string, metadata map[string]string) ([]byte, error) {
	if s.keySet == nil {
		return nil, ErrMissingRSADisableToken
	}
	// use first key to sign token
	rsaKey, ok := s.keySet.Key(0)
	if !ok {
		return nil, errors.New("missing rsa key to generate token")
	}

	// shield generated token has an extra custom claim
	// used to identify which public key to use to verify the token
	metadata[GeneratedClaimKey] = GeneratedClaimValue
	return utils.BuildToken(rsaKey, s.issuer, userID, tokenValidity, metadata)
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
