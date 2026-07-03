package user

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"net/mail"
	"reflect"
	"strings"
	"time"

	"github.com/raystack/salt/rql"

	"github.com/raystack/frontier/pkg/utils"

	"github.com/raystack/frontier/core/auditrecord/models"
	"github.com/raystack/frontier/core/avatar"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	pkgAuditRecord "github.com/raystack/frontier/pkg/auditrecord"
	"github.com/raystack/frontier/pkg/errors"
	"github.com/raystack/frontier/pkg/str"
)

const CSVContentType = "text/csv"

type RelationService interface {
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	BatchCheckPermission(ctx context.Context, relations []relation.Relation) ([]relation.CheckPair, error)
	Delete(ctx context.Context, rel relation.Relation) error
	LookupSubjects(ctx context.Context, rel relation.Relation) ([]string, error)
	LookupResources(ctx context.Context, rel relation.Relation) ([]string, error)
}

type SessionService interface {
	DeleteByUserID(ctx context.Context, userID string) error
}

type AuditRecordRepository interface {
	Create(ctx context.Context, auditRecord models.AuditRecord) (models.AuditRecord, error)
}

type Service struct {
	repository            Repository
	relationService       RelationService
	sessionService        SessionService
	auditRecordRepository AuditRecordRepository
	avatarConfig          avatar.Config
	Now                   func() time.Time
}

func NewService(repository Repository, relationRepo RelationService,
	sessionService SessionService, auditRecordRepository AuditRecordRepository, avatarConfig avatar.Config) *Service {
	return &Service{
		repository:            repository,
		relationService:       relationRepo,
		sessionService:        sessionService,
		auditRecordRepository: auditRecordRepository,
		avatarConfig:          avatarConfig,
		Now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

// GetByID email or slug
func (s Service) GetByID(ctx context.Context, id string) (User, error) {
	if isValidEmail(id) {
		return s.GetByEmail(ctx, id)
	}
	if utils.IsValidUUID(id) {
		return s.repository.GetByID(ctx, id)
	}
	return s.repository.GetByName(ctx, strings.ToLower(id))
}

func (s Service) GetByIDs(ctx context.Context, userIDs []string) ([]User, error) {
	return s.repository.GetByIDs(ctx, userIDs)
}

func (s Service) GetByEmail(ctx context.Context, email string) (User, error) {
	email = strings.ToLower(email)
	return s.repository.GetByEmail(ctx, email)
}

func (s Service) Create(ctx context.Context, user User) (User, error) {
	if err := avatar.Validate(user.Avatar, s.avatarConfig); err != nil {
		return User{}, err
	}
	return s.repository.Create(ctx, User{
		Name:     strings.ToLower(user.Name),
		Email:    strings.ToLower(user.Email),
		State:    Enabled,
		Avatar:   user.Avatar,
		Title:    user.Title,
		Metadata: user.Metadata,
	})
}

func (s Service) List(ctx context.Context, flt Filter) ([]User, error) {
	// state gets filtered in db
	return s.repository.List(ctx, flt)
}

// Update by user uuid, email or slug
// Note(kushsharma): we don't actually update email field of the user, if we want to support it
// one security concern is that we need to ensure users can't misuse it to takeover
// invitations created for other users.
func (s Service) Update(ctx context.Context, toUpdate User) (User, error) {
	if err := avatar.Validate(toUpdate.Avatar, s.avatarConfig); err != nil {
		return User{}, err
	}
	id := toUpdate.ID
	toUpdate.Email = strings.ToLower(toUpdate.Email)
	toUpdate.Name = strings.ToLower(toUpdate.Name)
	if isValidEmail(id) {
		return s.UpdateByEmail(ctx, toUpdate)
	}
	if utils.IsValidUUID(id) {
		return s.repository.UpdateByID(ctx, toUpdate)
	}
	return s.repository.UpdateByName(ctx, toUpdate)
}

func (s Service) UpdateByEmail(ctx context.Context, toUpdate User) (User, error) {
	toUpdate.Email = strings.ToLower(toUpdate.Email)
	toUpdate.Name = strings.ToLower(toUpdate.Name)
	return s.repository.UpdateByEmail(ctx, toUpdate)
}

func (s Service) Enable(ctx context.Context, id string) error {
	if !utils.IsValidUUID(id) {
		return ErrInvalidID
	}
	return s.repository.SetState(ctx, id, Enabled)
}

// Disable is a reversible soft-stop: it flips the user's state to Disabled and
// soft-deletes active sessions, but leaves SpiceDB relations in place so Enable
// can restore the user's access exactly as it was. Disable is NOT a revocation —
// tearing down the tuples is Delete's job (see core/deleter).
//
// All authenticated access stops while a user is disabled. Session cookies are
// rejected because their rows are soft-deleted (and Session.IsValid filters
// DeletedAt). PAT/JWT/client-credential auth all resolve the principal via
// user.Service.GetByID, whose repository filters out users in the Disabled
// state — so those flows fail with ErrNotExist for the same user.
func (s Service) Disable(ctx context.Context, id string) error {
	if !utils.IsValidUUID(id) {
		return ErrInvalidID
	}
	if err := s.repository.SetState(ctx, id, Disabled); err != nil {
		return err
	}
	return s.sessionService.DeleteByUserID(ctx, id)
}

// Delete by user uuid
// don't call this directly, use cascade deleter
func (s Service) Delete(ctx context.Context, id string) error {
	if err := s.relationService.Delete(ctx, relation.Relation{Subject: relation.Subject{
		ID:        id,
		Namespace: schema.UserPrincipal,
	}}); err != nil {
		return err
	}
	return s.repository.Delete(ctx, id)
}

// Sudo add platform permissions to user
func (s Service) Sudo(ctx context.Context, id string, relationName string) error {
	currentUser, err := s.GetByID(ctx, id)
	if errors.Is(err, ErrNotExist) {
		if isValidEmail(id) {
			// create a new user
			currentUser, err = s.Create(ctx, User{
				Email: id,
				Name:  str.GenerateUserSlug(id),
			})
			if err != nil {
				return err
			}
		} else {
			// skip
			return nil
		}
	}
	if err != nil {
		return err
	}

	// validate the requested platform relation
	switch relationName {
	case schema.MemberRelationName, schema.AdminRelationName:
	default:
		return fmt.Errorf("invalid relation name, possible options are: %s, %s", schema.MemberRelationName, schema.AdminRelationName)
	}

	// Act on the exact relation, not the permission it grants: admin and member both
	// grant `check`, so a permission check would skip adding member to an existing
	// admin and break the admin->member downgrade. Safe to run again.
	if ok, err := s.IsSudo(ctx, currentUser.ID, relationName); err != nil {
		return err
	} else if ok {
		return nil
	}

	_, err = s.relationService.Create(ctx, relation.Relation{
		Object: relation.Object{
			ID:        schema.PlatformID,
			Namespace: schema.PlatformNamespace,
		},
		Subject: relation.Subject{
			ID:        currentUser.ID,
			Namespace: schema.UserPrincipal,
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

// UnSudo removes a platform relation (admin or member) from a user.
// It removes the exact relation requested — an `admin` relation can now actually
// be stripped. Both admin and member grants/removals are audited.
func (s Service) UnSudo(ctx context.Context, id, relationName string) error {
	switch relationName {
	case schema.AdminRelationName, schema.MemberRelationName:
	default:
		return fmt.Errorf("invalid relation name, possible options are: %s, %s", schema.MemberRelationName, schema.AdminRelationName)
	}

	currentUser, err := s.GetByID(ctx, id)
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
			Namespace: schema.UserPrincipal,
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

// recordPlatformAuditRecord logs a platform admin/member grant or revoke. Actor is
// left empty; the repository fills it from context (caller, or system actor at boot).
func (s Service) recordPlatformAuditRecord(ctx context.Context, u User, event pkgAuditRecord.Event, relationName string) error {
	_, err := s.auditRecordRepository.Create(ctx, models.AuditRecord{
		Event: event,
		Resource: models.Resource{
			ID:   schema.PlatformID,
			Type: pkgAuditRecord.PlatformType,
			Name: schema.PlatformID,
		},
		Target: &models.Target{
			ID:   u.ID,
			Type: pkgAuditRecord.UserType,
			Name: u.Name,
		},
		OrgID:      schema.PlatformOrgID.String(),
		OccurredAt: s.Now(),
		Metadata:   map[string]any{"relation": relationName},
	})
	return err
}

// IsSudo checks platform permissions.
// Platform permissions are:
// - superuser
// - check
func (s Service) IsSudo(ctx context.Context, id string, permissionName string) (bool, error) {
	status, err := s.IsSudos(ctx, []string{id}, permissionName)
	if err != nil {
		return false, err
	}
	return len(status) > 0, nil
}

func (s Service) IsSudos(ctx context.Context, ids []string, permissionName string) ([]relation.Relation, error) {
	relations := utils.Map(ids, func(id string) relation.Relation {
		return relation.Relation{
			Subject: relation.Subject{
				ID:        id,
				Namespace: schema.UserPrincipal,
			},
			Object: relation.Object{
				ID:        schema.PlatformID,
				Namespace: schema.PlatformNamespace,
			},
			RelationName: permissionName,
		}
	})
	statusForIDs, err := s.relationService.BatchCheckPermission(ctx, relations)
	if err != nil {
		return nil, err
	}

	successChecks := utils.Filter(statusForIDs, func(pair relation.CheckPair) bool {
		return pair.Status
	})
	return utils.Map(successChecks, func(pair relation.CheckPair) relation.Relation {
		return pair.Relation
	}), nil
}

func (s Service) Search(ctx context.Context, rql *rql.Query) (SearchUserResponse, error) {
	return s.repository.Search(ctx, rql)
}

func isValidEmail(str string) bool {
	_, err := mail.ParseAddress(str)
	return err == nil
}

// IsValidEmail checks if the string is a valid email address
func IsValidEmail(str string) bool {
	return isValidEmail(str)
}

type CSVExport struct {
	UserID    string `csv:"User ID"`
	Name      string `csv:"Name"`
	Title     string `csv:"Title"`
	Email     string `csv:"Email"`
	State     string `csv:"State"`
	CreatedAt string `csv:"Joined At"`
}

func (s Service) Export(ctx context.Context) ([]byte, string, error) {
	userData, err := s.repository.Search(ctx, &rql.Query{})
	if err != nil {
		return nil, "", err
	}

	if len(userData.Users) == 0 {
		return nil, "", ErrNoUsersFound
	}

	// Create a buffer to write CSV data
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write headers
	csvExport := NewCSVExport(userData.Users[0])
	headers := csvExport.GetHeaders()
	if err := writer.Write(headers); err != nil {
		return nil, "", err
	}

	// Write data rows
	for _, user := range userData.Users {
		csvExport := NewCSVExport(user)
		if err := writer.Write(csvExport.ToRow()); err != nil {
			return nil, "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), CSVContentType, nil
}

// NewCSVExport converts User to CSVExport
func NewCSVExport(user User) CSVExport {
	return CSVExport{
		UserID:    user.ID,
		Name:      user.Name,
		Title:     user.Title,
		Email:     user.Email,
		State:     string(user.State),
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}
}

// GetHeaders returns the CSV headers based on struct tags
func (c CSVExport) GetHeaders() []string {
	t := reflect.TypeOf(c)
	headers := make([]string, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if tag := field.Tag.Get("csv"); tag != "" {
			headers[i] = tag
		} else {
			headers[i] = field.Name
		}
	}

	return headers
}

// ToRow converts the struct to a string slice for CSV writing
func (c CSVExport) ToRow() []string {
	v := reflect.ValueOf(c)
	row := make([]string, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		row[i] = fmt.Sprint(v.Field(i).Interface())
	}

	return row
}
