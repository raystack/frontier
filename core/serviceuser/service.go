package serviceuser

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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

	// TODO(kushsharma): delete all of serviceuser relationships
	// before deleting the serviceuser
	return s.repo.Delete(ctx, id)
}

func (s Service) ListKeys(ctx context.Context, serviceUserID string) ([]Credential, error) {
	return s.credRepo.List(ctx, Filter{
		ServiceUserID: serviceUserID,
	})
}

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
	return s.credRepo.Get(ctx, credID)
}

func (s Service) DeleteKey(ctx context.Context, credID string) error {
	return s.credRepo.Delete(ctx, credID)
}

func (s Service) CreateSecret(ctx context.Context, serviceUserID string) (Credential, error) {
	//TODO implement me
	panic("implement me")
}

func (s Service) GetSecret(ctx context.Context, serviceUserID string, credID string) (Credential, error) {
	//TODO implement me
	panic("implement me")
}

func (s Service) DeleteSecret(ctx context.Context, serviceUserID string, credID string) error {
	//TODO implement me
	panic("implement me")
}

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
