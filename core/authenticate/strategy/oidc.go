package strategy

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type OIDC struct {
	Client   *http.Client
	config   *oauth2.Config
	provider *oidc.Provider
}

type UserInfo struct {
	Name  string
	Email string
}

func NewRelyingPartyOIDC(clientId string, clientSecret string, redirectUrl string) *OIDC {
	return &OIDC{
		Client: http.DefaultClient,
		config: &oauth2.Config{
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
			ClientID:     clientId,
			ClientSecret: clientSecret,
			RedirectURL:  redirectUrl,
		},
	}
}

func (g *OIDC) Init(ctx context.Context, issuer string) (*OIDC, error) {
	var err error
	if g.provider, err = oidc.NewProvider(ctx, issuer); err != nil {
		return nil, err
	}
	g.config.Endpoint = g.provider.Endpoint()
	return g, err
}

func (g *OIDC) AuthURL(state string) (url string, nonce string, err error) {
	nonce, err = generateNonce()
	if err != nil {
		return
	}
	url = g.config.AuthCodeURL(state, oidc.Nonce(nonce))
	return
}

// Token exchange auth code with token and verifies that an *oauth2.Token is a valid *oidc.IDToken
// it matches nonce from *oidc.IDToken with flow nonce
func (g *OIDC) Token(ctx context.Context, code string, nonce string) (*oauth2.Token, error) {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, g.Client)
	authToken, err := g.config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	rawIDToken, ok := authToken.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("missing id_token")
	}

	idToken, err := g.provider.Verifier(&oidc.Config{
		ClientID: g.config.ClientID,
	}).Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("invalid id_token, validation failed: %w", err)
	}

	if idToken.Nonce != nonce {
		return nil, errors.New("invalid id_token, validation failed")
	}

	return authToken, nil
}

func (g *OIDC) GetUser(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	baseUser, err := g.provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
	if err != nil {
		return nil, err
	}
	var userClaims map[string]any
	if err = baseUser.Claims(&userClaims); err != nil {
		return nil, err
	}

	if baseUser.Email == "" {
		return nil, errors.New("invalid email")
	}
	user := &UserInfo{
		Name:  baseUser.Profile,
		Email: baseUser.Email,
	}

	// try few fields which are possibly contain name of the user
	if name, ok := userClaims["name"].(string); ok {
		user.Name = name
	}
	if name, ok := userClaims["full_name"].(string); ok {
		user.Name = name
	}
	return user, nil
}

func EmbedFlowInOIDCState(param string) (string, error) {
	randBytes := make([]byte, 8)
	if _, err := io.ReadFull(rand.Reader, randBytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf("%s::%s", param, randBytes))), nil
}

func ExtractFlowFromOIDCState(state string) (string, error) {
	stateBytes, err := base64.RawURLEncoding.DecodeString(state)
	if err != nil {
		return "", err
	}
	stateParts := strings.Split(string(stateBytes), "::")
	flowID := stateParts[0] // first part is flow id
	return flowID, nil
}

func generateNonce() (string, error) {
	nonceBytes := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, nonceBytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(nonceBytes), nil
}
