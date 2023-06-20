package serviceuser

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/gtank/cryptopasta"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/raystack/shield/core/relation"
	"github.com/raystack/shield/internal/bootstrap/schema"
	"github.com/raystack/shield/pkg/utils"
)

type Repository interface {
	List(ctx context.Context, flt Filter) ([]ServiceUser, error)
	Create(ctx context.Context, serviceUser ServiceUser) (ServiceUser, error)
	Get(ctx context.Context, id string) (ServiceUser, error)
	Delete(ctx context.Context, id string) error
}

type CredentialRepository interface {
	List(ctx context.Context, flt Filter) ([]Credential, error)
	Create(ctx context.Context, credential Credential) (Credential, error)
	Get(ctx context.Context, id string) (Credential, error)
	Delete(ctx context.Context, id string) error
}

type RelationService interface {
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	Delete(ctx context.Context, rel relation.Relation) error
}

type Service struct {
	repo       Repository
	credRepo   CredentialRepository
	relService RelationService
}

func NewService(repo Repository, credRepo CredentialRepository, relService RelationService) *Service {
	return &Service{
		repo:       repo,
		credRepo:   credRepo,
		relService: relService,
	}
}

func (s Service) List(ctx context.Context, flt Filter) ([]ServiceUser, error) {
	return s.repo.List(ctx, flt)
}

func (s Service) Create(ctx context.Context, serviceUser ServiceUser) (ServiceUser, error) {
	createdSvUser, err := s.repo.Create(ctx, serviceUser)
	if err != nil {
		return ServiceUser{}, err
	}

	// attach the serviceuser to the organization
	_, err = s.relService.Create(ctx, relation.Relation{
		Object: relation.Object{
			ID:        createdSvUser.ID,
			Namespace: schema.ServiceUserPrincipal,
		},
		Subject: relation.Subject{
			ID:        createdSvUser.OrgID,
			Namespace: schema.OrganizationNamespace,
		},
		RelationName: schema.OrganizationRelationName,
	})
	if err != nil {
		return ServiceUser{}, err
	}
	_, err = s.relService.Create(ctx, relation.Relation{
		Object: relation.Object{
			ID:        createdSvUser.ID,
			Namespace: schema.ServiceUserPrincipal,
		},
		Subject: relation.Subject{
			ID:        createdSvUser.OrgID,
			Namespace: schema.OrganizationNamespace,
		},
		RelationName: schema.OrganizationRelationName,
	})
	if err != nil {
		return ServiceUser{}, err
	}
	return createdSvUser, nil
}

func (s Service) Get(ctx context.Context, id string) (ServiceUser, error) {
	return s.repo.Get(ctx, id)
}

func (s Service) Delete(ctx context.Context, id string) error {
	// delete all of its credentials then delete the service account
	creds, err := s.credRepo.List(ctx, Filter{
		ServiceUserID: id,
	})
	if err != nil {
		return err
	}
	for _, cred := range creds {
		if err := s.credRepo.Delete(ctx, cred.ID); err != nil {
			return err
		}
	}

	// delete all of serviceuser relationships
	// before deleting the serviceuser
	if err := s.relService.Delete(ctx, relation.Relation{
		Object: relation.Object{
			ID:        id,
			Namespace: schema.ServiceUserPrincipal,
		},
	}); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s Service) ListKeys(ctx context.Context, serviceUserID string) ([]Credential, error) {
	return s.credRepo.List(ctx, Filter{
		ServiceUserID: serviceUserID,
		IsKey:         true,
	})
}

// CreateKey creates a key pair for the service user
func (s Service) CreateKey(ctx context.Context, credential Credential) (Credential, error) {
	credential.ID = uuid.New().String()

	// generate public/private key pair
	newJWK, err := utils.CreateJWKWithKID(credential.ID)
	if err != nil {
		return Credential{}, fmt.Errorf("failed to create key pair: %w", err)
	}
	jwkPEM, err := jwk.Pem(newJWK)
	if err != nil {
		return Credential{}, fmt.Errorf("failed to convert jwk to pem: %w", err)
	}

	pubKey, err := newJWK.PublicKey()
	if err != nil {
		return Credential{}, err
	}

	// using single key for now
	publicKeySet := jwk.NewSet()
	publicKeySet.AddKey(pubKey)
	credential.PublicKey = publicKeySet

	// save public key in database
	createdCred, err := s.credRepo.Create(ctx, credential)
	if err != nil {
		return Credential{}, err
	}
	createdCred.PrivateKey = jwkPEM
	return createdCred, nil
}

func (s Service) GetKey(ctx context.Context, credID string) (Credential, error) {
	cred, err := s.credRepo.Get(ctx, credID)
	if err != nil {
		return Credential{}, err
	}
	if len(cred.SecretHash) > 0 || cred.PublicKey == nil || cred.PublicKey.Len() == 0 {
		return Credential{}, ErrCredNotExist
	}
	return cred, err
}

func (s Service) DeleteKey(ctx context.Context, credID string) error {
	return s.credRepo.Delete(ctx, credID)
}

// CreateSecret creates a secret for the service user
func (s Service) CreateSecret(ctx context.Context, credential Credential) (Secret, error) {
	// generate a random secret
	secretBytes := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, secretBytes); err != nil {
		return Secret{}, err
	}
	secretString := base64.RawStdEncoding.EncodeToString(secretBytes)
	if sHash, err := cryptopasta.HashPassword([]byte(secretString)); err != nil {
		return Secret{}, err
	} else {
		credential.SecretHash = sHash
	}

	createdCred, err := s.credRepo.Create(ctx, credential)
	if err != nil {
		return Secret{}, err
	}
	return Secret{
		ID:        createdCred.ID,
		Value:     []byte(secretString),
		CreatedAt: createdCred.CreatedAt,
	}, nil
}

func (s Service) DeleteSecret(ctx context.Context, credID string) error {
	return s.DeleteKey(ctx, credID)
}

func (s Service) ListSecret(ctx context.Context, serviceUserID string) ([]Credential, error) {
	return s.credRepo.List(ctx, Filter{
		ServiceUserID: serviceUserID,
		IsSecret:      true,
	})
}

// GetBySecret matches the secret with the secret hash stored in the database of the service user
// and if the secret matches, returns the service user
func (s Service) GetBySecret(ctx context.Context, credID string, credSecret string) (ServiceUser, error) {
	cred, err := s.credRepo.Get(ctx, credID)
	if err != nil {
		return ServiceUser{}, err
	}
	if len(cred.SecretHash) <= 0 {
		return ServiceUser{}, ErrInvalidCred
	}
	if err := cryptopasta.CheckPasswordHash(cred.SecretHash, []byte(credSecret)); err != nil {
		return ServiceUser{}, ErrInvalidCred
	}
	return s.repo.Get(ctx, cred.ServiceUserID)
}

// GetByToken returns the service user by verifying the token
func (s Service) GetByToken(ctx context.Context, token string) (ServiceUser, error) {
	insecureToken, err := jwt.ParseInsecure([]byte(token))
	if err != nil {
		return ServiceUser{}, fmt.Errorf("invalid serviceuser token: %w", err)
	}
	tokenKID, ok := insecureToken.Get(jwk.KeyIDKey)
	if !ok {
		return ServiceUser{}, fmt.Errorf("invalid key id from token")
	}
	cred, err := s.credRepo.Get(ctx, tokenKID.(string))
	if err != nil {
		return ServiceUser{}, fmt.Errorf("credential invalid of kid %s: %w", tokenKID.(string), err)
	}

	// verify token
	_, err = jwt.Parse([]byte(token), jwt.WithKeySet(cred.PublicKey))
	if err != nil {
		return ServiceUser{}, fmt.Errorf("invalid serviceuser token: %w", err)
	}
	return s.repo.Get(ctx, cred.ServiceUserID)
}
