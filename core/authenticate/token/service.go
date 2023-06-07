package token

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/odpf/shield/pkg/server/consts"
	"google.golang.org/grpc/metadata"
)

var (
	ErrMissingRSADisableToken = errors.New("rsa key missing in config, generate and pass file path")
	ErrInvalidToken           = errors.New("failed to verify a valid token")
	ErrNoToken                = errors.New("no token")
)

const (
	// TODO(kushsharma): should we expose this in config?
	tokenValidity = time.Hour * 24 * 7
)

type Service struct {
	keySet       jwk.Set
	publicKeySet jwk.Set
	issuer       string
}

func NewService(keySet jwk.Set, issuer string) Service {
	publicKeySet := jwk.NewSet()
	if keySet != nil {
		// convert private to public
		for iter := keySet.Keys(context.Background()); iter.Next(context.Background()); {
			pair := iter.Pair()
			key := pair.Value.(jwk.Key)

			pubKey, err := key.PublicKey()
			if err != nil {
				panic(fmt.Errorf("failed to generate public key from private rsa: %w", err))
			}
			publicKeySet.AddKey(pubKey)
		}
	}

	return Service{
		keySet:       keySet,
		issuer:       issuer,
		publicKeySet: publicKeySet,
	}
}

func (s Service) Build(ctx context.Context, userID string, metadata map[string]string) ([]byte, error) {
	if s.keySet == nil {
		return nil, ErrMissingRSADisableToken
	}

	// use first key to sign token
	rsaKey, ok := s.keySet.Key(0)
	if !ok {
		return nil, errors.New("missing rsa key to generate token")
	}

	body := jwt.NewBuilder().
		Issuer(s.issuer).
		IssuedAt(time.Now().UTC()).
		NotBefore(time.Now().UTC()).
		Expiration(time.Now().UTC().Add(tokenValidity)).
		JwtID(uuid.New().String()).
		Subject(userID)
	for claimKey, claimVal := range metadata {
		body = body.Claim(claimKey, claimVal)
	}

	tok, err := body.Build()
	if err != nil {
		return nil, err
	}
	return jwt.Sign(tok, jwt.WithKey(jwa.RS256, rsaKey))
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

func (s Service) ParseFromContext(ctx context.Context) (string, map[string]any, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", nil, ErrNoToken
	}

	tokenHeaders := md.Get(consts.UserTokenGatewayKey)
	if len(tokenHeaders) == 0 || len(tokenHeaders[0]) == 0 {
		return "", nil, ErrNoToken
	}
	return s.Parse(ctx, []byte(tokenHeaders[0]))
}
