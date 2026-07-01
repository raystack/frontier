package serviceuser

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"golang.org/x/crypto/sha3"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/raystack/frontier/core/auditrecord/models"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	pkgAuditRecord "github.com/raystack/frontier/pkg/auditrecord"
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
	CheckPermission(ctx context.Context, rel relation.Relation) (bool, error)
	BatchCheckPermission(ctx context.Context, rel []relation.Relation) ([]relation.CheckPair, error)
}

type MembershipService interface {
	AddOrganizationMember(ctx context.Context, orgID, principalID, principalType, roleID string) error
	RemoveOrganizationMember(ctx context.Context, orgID, principalID, principalType string) error
	ListPrincipalIDsByResource(ctx context.Context, resourceID, resourceType, principalType string) ([]string, error)
}

type AuditRecordRepository interface {
	Create(ctx context.Context, auditRecord models.AuditRecord) (models.AuditRecord, error)
}

type Service struct {
	log                   *slog.Logger
	repo                  Repository
	credRepo              CredentialRepository
	relationService       RelationService
	membershipService     MembershipService
	auditRecordRepository AuditRecordRepository
}

func NewService(logger *slog.Logger, repo Repository, credRepo CredentialRepository, relService RelationService, auditRecordRepository AuditRecordRepository) *Service {
	return &Service{
		log:                   logger,
		repo:                  repo,
		credRepo:              credRepo,
		relationService:       relService,
		auditRecordRepository: auditRecordRepository,
	}
}

// SetMembershipService sets the membership dependency after construction to break
// the circular init order between serviceuser and membership services.
func (s *Service) SetMembershipService(ms MembershipService) {
	s.membershipService = ms
}

func (s Service) List(ctx context.Context, flt Filter) ([]ServiceUser, error) {
	if flt.OrgID == "" && len(flt.ServiceUserIDs) == 0 && flt.ServiceUserID == "" {
		return nil, ErrInvalidID
	}
	return s.repo.List(ctx, flt)
}

func (s Service) ListAll(ctx context.Context) ([]ServiceUser, error) {
	// ListAll allows listing all service users without any filtering
	// This is intended for admin usage to get all service users across all organizations
	return s.repo.List(ctx, Filter{})
}

func (s Service) Create(ctx context.Context, serviceUser ServiceUser) (ServiceUser, error) {
	createdSU, err := s.repo.Create(ctx, serviceUser)
	if err != nil {
		return ServiceUser{}, err
	}

	// add service user as org member with default viewer role
	// creates policy + org#member relation + serviceuser#org identity link
	if err := s.membershipService.AddOrganizationMember(ctx, serviceUser.OrgID, createdSU.ID, schema.ServiceUserPrincipal, schema.RoleOrganizationViewer); err != nil {
		// rollback: delete the orphan SU row to avoid accumulating dead records
		if deleteErr := s.repo.Delete(ctx, createdSU.ID); deleteErr != nil {
			s.log.ErrorContext(ctx, "orphan serviceuser: membership setup failed and rollback delete also failed, manual cleanup needed",
				"serviceuser_id", createdSU.ID,
				"org_id", serviceUser.OrgID,
				"membership_error", err,
				"delete_error", deleteErr,
			)
		}
		return ServiceUser{}, fmt.Errorf("add org membership: %w", err)
	}

	return createdSU, nil
}

func (s Service) Get(ctx context.Context, id string) (ServiceUser, error) {
	if !utils.IsValidUUID(id) {
		return ServiceUser{}, ErrInvalidID
	}
	return s.repo.GetByID(ctx, id)
}

func (s Service) GetByIDs(ctx context.Context, ids []string) ([]ServiceUser, error) {
	return s.repo.GetByIDs(ctx, ids)
}

func (s Service) ListByOrg(ctx context.Context, orgID string) ([]ServiceUser, error) {
	userIDs, err := s.membershipService.ListPrincipalIDsByResource(ctx, orgID, schema.OrganizationNamespace, schema.ServiceUserPrincipal)
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
	// fetch SU to get org ID for membership cleanup
	su, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// delete all of its credentials
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

	// remove org membership (policies at org/project/group level + org relations)
	// best-effort: log and continue on failure — leaving a half-deleted SU is worse than a leaked policy
	if err := s.membershipService.RemoveOrganizationMember(ctx, su.OrgID, id, schema.ServiceUserPrincipal); err != nil {
		s.log.ErrorContext(ctx, "failed to remove org membership during serviceuser delete, policies may be leaked",
			"serviceuser_id", id,
			"org_id", su.OrgID,
			"error", err,
		)
	}

	// SU may appear as Subject (e.g. platform sudo) or as Object (e.g. the
	// serviceuser#org identity link). Sweep both sides so deletion is symmetric
	// with creation regardless of whether the membership cascade ran above.
	if err := s.relationService.Delete(ctx, relation.Relation{
		Subject: relation.Subject{
			ID:        id,
			Namespace: schema.ServiceUserPrincipal,
		},
	}); err != nil {
		return err
	}
	if err := s.relationService.Delete(ctx, relation.Relation{
		Object: relation.Object{
			ID:        id,
			Namespace: schema.ServiceUserPrincipal,
		},
	}); err != nil && !errors.Is(err, relation.ErrNotExist) {
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

	// validate the requested platform relation
	switch relationName {
	case schema.MemberRelationName, schema.AdminRelationName:
	default:
		return fmt.Errorf("invalid relation name, possible options are: %s, %s", schema.MemberRelationName, schema.AdminRelationName)
	}

	// Check the exact relation, not the permission it grants. Running this again
	// is safe. Both admin and member grant `check`, so checking the permission
	// would skip creating the member relation for someone who is already an admin.
	// That breaks a downgrade (Sudo member, then UnSudo admin), because UnSudo
	// works at the relation level. Keep the two the same.
	if ok, err := s.IsSudo(ctx, currentUser.ID, relationName); err != nil {
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
	if err != nil {
		return err
	}

	// audit the grant for both admin and member relations
	event := pkgAuditRecord.PlatformAdminAddedEvent
	if relationName == schema.MemberRelationName {
		event = pkgAuditRecord.PlatformMemberAddedEvent
	}
	return s.recordPlatformAuditRecord(ctx, currentUser, event, relationName)
}

// UnSudo removes a platform relation (admin or member) from a service user.
// It removes the exact relation requested — an `admin` relation can now actually
// be stripped. Both admin and member grants/removals are audited.
func (s Service) UnSudo(ctx context.Context, id, relationName string) error {
	switch relationName {
	case schema.AdminRelationName, schema.MemberRelationName:
	default:
		return fmt.Errorf("invalid relation name, possible options are: %s, %s", schema.MemberRelationName, schema.AdminRelationName)
	}

	currentUser, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	// Only act (and audit) when the specific relation actually exists, so the
	// revoke event reflects a real state change. Checking the relation directly
	// is precise for both admin and member.
	present, err := s.IsSudo(ctx, currentUser.ID, relationName)
	if err != nil {
		return err
	}
	if !present {
		return nil
	}

	// unmark su
	if err := s.relationService.Delete(ctx, relation.Relation{
		Object: relation.Object{
			ID:        schema.PlatformID,
			Namespace: schema.PlatformNamespace,
		},
		Subject: relation.Subject{
			ID:        currentUser.ID,
			Namespace: schema.ServiceUserPrincipal,
		},
		RelationName: relationName,
	}); err != nil {
		return err
	}

	event := pkgAuditRecord.PlatformAdminRemovedEvent
	if relationName == schema.MemberRelationName {
		event = pkgAuditRecord.PlatformMemberRemovedEvent
	}
	return s.recordPlatformAuditRecord(ctx, currentUser, event, relationName)
}

// recordPlatformAuditRecord writes an audit record when a platform relation
// (admin or member) is granted or removed on a service user. Actor is left empty,
// so the repository fills it in from the context (the caller, or the system actor
// at boot).
func (s Service) recordPlatformAuditRecord(ctx context.Context, su ServiceUser, event pkgAuditRecord.Event, relationName string) error {
	_, err := s.auditRecordRepository.Create(ctx, models.AuditRecord{
		Event: event,
		Resource: models.Resource{
			ID:   schema.PlatformID,
			Type: pkgAuditRecord.PlatformType,
			Name: schema.PlatformID,
		},
		Target: &models.Target{
			ID:   su.ID,
			Type: pkgAuditRecord.ServiceUserType,
			Name: su.Title,
		},
		OrgID:      schema.PlatformOrgID.String(),
		OccurredAt: time.Now().UTC(),
		Metadata:   map[string]any{"relation": relationName},
	})
	return err
}
