package serviceuser

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"

	"golang.org/x/crypto/sha3"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/utils"
)

type Repository interface {
	List(ctx context.Context, flt Filter) ([]ServiceUser, error)
	Create(ctx context.Context, serviceUser ServiceUser) (ServiceUser, error)
	GetByID(ctx context.Context, id string) (ServiceUser, error)
	GetByIDs(ctx context.Context, id []string) ([]ServiceUser, error)
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
	LookupSubjects(ctx context.Context, rel relation.Relation) ([]string, error)
	CheckPermission(ctx context.Context, rel relation.Relation) (bool, error)
	BatchCheckPermission(ctx context.Context, rel []relation.Relation) ([]relation.CheckPair, error)
}

type Service struct {
	repo            Repository
	credRepo        CredentialRepository
	relationService RelationService
}

func NewService(repo Repository, credRepo CredentialRepository, relService RelationService) *Service {
	return &Service{
		repo:            repo,
		credRepo:        credRepo,
		relationService: relService,
	}
}

func (s Service) List(ctx context.Context, flt Filter) ([]ServiceUser, error) {
	if flt.OrgID == "" && len(flt.ServiceUserIDs) == 0 && flt.ServiceUserID == "" {
		return nil, ErrInvalidID
	}
	return s.repo.List(ctx, flt)
}

func (s Service) Create(ctx context.Context, serviceUser ServiceUser) (ServiceUser, error) {
	createdSU, err := s.repo.Create(ctx, serviceUser)
	if err != nil {
		return ServiceUser{}, err
	}

	// attach service user to organization
	_, err = s.relationService.Create(ctx, relation.Relation{
		Object: relation.Object{
			ID:        serviceUser.OrgID,
			Namespace: schema.OrganizationNamespace,
		},
		Subject: relation.Subject{
			ID:        createdSU.ID,
			Namespace: schema.ServiceUserPrincipal,
		},
		RelationName: schema.MemberRelationName,
	})
	if err != nil {
		return ServiceUser{}, err
	}
	_, err = s.relationService.Create(ctx, relation.Relation{
		Object: relation.Object{
			ID:        createdSU.ID,
			Namespace: schema.ServiceUserPrincipal,
		},
		Subject: relation.Subject{
			ID:        serviceUser.OrgID,
			Namespace: schema.OrganizationNamespace,
		},
		RelationName: schema.OrganizationRelationName,
	})
	if err != nil {
		return ServiceUser{}, err
	}

	if len(serviceUser.CreatedByUser) > 0 {
		// TODO: write authz tests that checks if the user who created the service user
		// has the permission to interact with the service user
		// attach user to service user who created it
		_, err = s.relationService.Create(ctx, relation.Relation{
			Object: relation.Object{
				ID:        createdSU.ID,
				Namespace: schema.ServiceUserPrincipal,
			},
			Subject: relation.Subject{
				ID:        serviceUser.CreatedByUser,
				Namespace: schema.UserPrincipal,
			},
			RelationName: schema.UserRelationName,
		})
		if err != nil {
			return ServiceUser{}, err
		}
	}

	return createdSU, nil
}

func (s Service) Get(ctx context.Context, id string) (ServiceUser, error) {
	return s.repo.GetByID(ctx, id)
}

func (s Service) GetByIDs(ctx context.Context, ids []string) ([]ServiceUser, error) {
	return s.repo.GetByIDs(ctx, ids)
}

func (s Service) ListByOrg(ctx context.Context, orgID string) ([]ServiceUser, error) {
	userIDs, err := s.relationService.LookupSubjects(ctx, relation.Relation{
		Object: relation.Object{
			ID:        orgID,
			Namespace: schema.OrganizationNamespace,
		},
		Subject: relation.Subject{
			Namespace: schema.ServiceUserPrincipal,
		},
		RelationName: schema.MembershipPermission,
	})
	if err != nil {
		return nil, err
	}
	if len(userIDs) == 0 {
		// no users
		return []ServiceUser{}, nil
	}
	return s.repo.GetByIDs(ctx, userIDs)
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
	if err := s.relationService.Delete(ctx, relation.Relation{
		Subject: relation.Subject{
			ID:        id,
			Namespace: schema.ServiceUserPrincipal,
		},
	}); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s Service) ListKeys(ctx context.Context, serviceUserID string) ([]Credential, error) {
	if serviceUserID == "" {
		return nil, ErrInvalidID
	}
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
	if err := publicKeySet.AddKey(pubKey); err != nil {
		return Credential{}, err
	}
	credential.PublicKey = publicKeySet

	// save public key in database
	credential.Type = JWTCredentialType
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
	if sHash, err := bcrypt.GenerateFromPassword([]byte(secretString), 14); err != nil {
		return Secret{}, err
	} else {
		credential.SecretHash = string(sHash)
	}

	credential.Type = ClientSecretCredentialType
	createdCred, err := s.credRepo.Create(ctx, credential)
	if err != nil {
		return Secret{}, err
	}
	return Secret{
		ID:        createdCred.ID,
		Title:     createdCred.Title,
		Value:     secretString,
		CreatedAt: createdCred.CreatedAt,
	}, nil
}

func (s Service) DeleteSecret(ctx context.Context, credID string) error {
	return s.DeleteKey(ctx, credID)
}

func (s Service) ListSecret(ctx context.Context, serviceUserID string) ([]Credential, error) {
	if serviceUserID == "" {
		return nil, ErrInvalidID
	}
	return s.credRepo.List(ctx, Filter{
		ServiceUserID: serviceUserID,
		IsSecret:      true,
	})
}

// CreateToken creates an opaque token for the service user
func (s Service) CreateToken(ctx context.Context, credential Credential) (Token, error) {
	// generate a random secret
	secretBytes := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, secretBytes); err != nil {
		return Token{}, err
	}

	// Hash the random bytes using SHA3-256
	hash := sha3.Sum256(secretBytes)
	credential.SecretHash = hex.EncodeToString(hash[:])

	credential.Type = OpaqueTokenCredentialType
	createdCred, err := s.credRepo.Create(ctx, credential)
	if err != nil {
		return Token{}, err
	}

	// encode cred val to hex bytes
	credVal := hex.EncodeToString(secretBytes)
	return Token{
		ID:        createdCred.ID,
		Title:     createdCred.Title,
		Value:     credVal,
		CreatedAt: createdCred.CreatedAt,
	}, nil
}

func (s Service) DeleteToken(ctx context.Context, credID string) error {
	return s.DeleteKey(ctx, credID)
}

func (s Service) ListToken(ctx context.Context, serviceUserID string) ([]Credential, error) {
	return s.ListSecret(ctx, serviceUserID)
}

// GetBySecret matches the secret with the secret hash stored in the database of the service user
// and if the secret matches, returns the service user
func (s Service) GetBySecret(ctx context.Context, credID string, reqSecret string) (ServiceUser, error) {
	cred, err := s.credRepo.Get(ctx, credID)
	if err != nil {
		return ServiceUser{}, err
	}
	if len(cred.SecretHash) <= 0 || len(reqSecret) <= 0 {
		return ServiceUser{}, ErrInvalidCred
	}
	if len(cred.Type) == 0 || cred.Type == ClientSecretCredentialType {
		if err := bcrypt.CompareHashAndPassword([]byte(cred.SecretHash), []byte(reqSecret)); err != nil {
			return ServiceUser{}, ErrInvalidCred
		}
	}
	if cred.Type == OpaqueTokenCredentialType {
		// decode the hex encoded secret
		decodedReqSecret := make([]byte, 32)
		if _, err := hex.Decode(decodedReqSecret, []byte(reqSecret)); err != nil {
			return ServiceUser{}, err
		}
		reqDigest := sha3.Sum256(decodedReqSecret)

		decodedCredSecret := make([]byte, 32)
		if _, err := hex.Decode(decodedCredSecret, []byte(cred.SecretHash)); err != nil {
			return ServiceUser{}, err
		}
		if subtle.ConstantTimeCompare(reqDigest[:], decodedCredSecret) == 0 {
			return ServiceUser{}, ErrInvalidCred
		}
	}
	return s.repo.GetByID(ctx, cred.ServiceUserID)
}

// GetByJWT returns the service user by verifying the token
func (s Service) GetByJWT(ctx context.Context, token string) (ServiceUser, error) {
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
	return s.repo.GetByID(ctx, cred.ServiceUserID)
}

// IsSudo checks platform permissions.
// Platform permissions are:
// - superuser
// - check
func (s Service) IsSudo(ctx context.Context, id string, permissionName string) (bool, error) {
	return s.relationService.CheckPermission(ctx, relation.Relation{
		Subject: relation.Subject{
			ID:        id,
			Namespace: schema.ServiceUserPrincipal,
		},
		Object: relation.Object{
			ID:        schema.PlatformID,
			Namespace: schema.PlatformNamespace,
		},
		RelationName: permissionName,
	})
}

// FilterSudos filters serviceusers which have superuser permissions and returns the remaining
func (s Service) FilterSudos(ctx context.Context, ids []string) ([]string, error) {
	relations := make([]relation.Relation, 0, len(ids))
	for _, id := range ids {
		rel := relation.Relation{
			Subject: relation.Subject{
				ID:        id,
				Namespace: schema.ServiceUserPrincipal,
			},
			Object: relation.Object{
				ID:        schema.PlatformID,
				Namespace: schema.PlatformNamespace,
			},
			RelationName: schema.PlatformSudoPermission,
		}
		relations = append(relations, rel)
	}
	checkPairs, err := s.relationService.BatchCheckPermission(ctx, relations)
	if err != nil {
		return nil, err
	}
	sudoIDs := make([]string, 0, len(checkPairs))
	for i, checkPair := range checkPairs {
		if !checkPair.Status {
			sudoIDs = append(sudoIDs, ids[i])
		}
	}
	return sudoIDs, nil
}

// Sudo add platform permissions to user
func (s Service) Sudo(ctx context.Context, id string, relationName string) error {
	currentUser, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	// check if already su
	permissionName := ""
	switch relationName {
	case schema.MemberRelationName:
		permissionName = schema.PlatformCheckPermission
	case schema.AdminRelationName:
		permissionName = schema.PlatformSudoPermission
	}
	if permissionName == "" {
		return fmt.Errorf("invalid relation name, possible options are: %s, %s", schema.MemberRelationName, schema.AdminRelationName)
	}

	if ok, err := s.IsSudo(ctx, currentUser.ID, permissionName); err != nil {
		return err
	} else if ok {
		return nil
	}

	// mark su
	_, err = s.relationService.Create(ctx, relation.Relation{
		Object: relation.Object{
			ID:        schema.PlatformID,
			Namespace: schema.PlatformNamespace,
		},
		Subject: relation.Subject{
			ID:        currentUser.ID,
			Namespace: schema.ServiceUserPrincipal,
		},
		RelationName: relationName,
	})
	return err
}

// UnSudo remove platform permissions to user
// only remove the 'member' relation if it exists
func (s Service) UnSudo(ctx context.Context, id string) error {
	currentUser, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	relationName := schema.MemberRelationName
	// to check if the user has member relation, we need to check if the user has `check`
	// permission on platform
	if ok, err := s.IsSudo(ctx, currentUser.ID, schema.PlatformCheckPermission); err != nil {
		return err
	} else if !ok {
		// not needed
		return nil
	}

	// unmark su
	err = s.relationService.Delete(ctx, relation.Relation{
		Object: relation.Object{
			ID:        schema.PlatformID,
			Namespace: schema.PlatformNamespace,
		},
		Subject: relation.Subject{
			ID:        currentUser.ID,
			Namespace: schema.ServiceUserPrincipal,
		},
		RelationName: relationName,
	})
	return err
}
