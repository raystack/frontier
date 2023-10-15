package strategy

import (
	"strings"

	"github.com/go-webauthn/webauthn/webauthn"
)

const (
	PasskeyAuthMethod   string = "passkey"
	PasskeyRegisterType string = "register"
	PasskeyLoginType    string = "login"
)

type UserData struct {
	Id          string
	Name        string
	DisplayName string
	Credentials []webauthn.Credential
}

func NewPassKeyUser(id string) *UserData {
	user := &UserData{}
	user.Id = id
	user.Name = extractUsername(id)
	user.DisplayName = extractUsername(id)
	return user
}

func NewPasskeyUserWithCredentials(id string, webAuthCredentialData []webauthn.Credential) *UserData {
	user := &UserData{}
	user.Id = id
	user.Name = extractUsername(id)
	user.DisplayName = extractUsername(id)
	user.Credentials = webAuthCredentialData
	return user
}

func (u *UserData) WebAuthnID() []byte {
	return []byte(u.Id)
}

func (u *UserData) WebAuthnName() string {
	return u.Name
}

func (u *UserData) WebAuthnDisplayName() string {
	return u.DisplayName
}
func (u *UserData) WebAuthnIcon() string {
	return ""
}
func (u *UserData) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}

func extractUsername(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) == 2 {
		return parts[0]
	}
	return ""
}
