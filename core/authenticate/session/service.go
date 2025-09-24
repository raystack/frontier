package session

import (
	"context"
	"errors"
	"time"

	"github.com/raystack/frontier/pkg/server/consts"

	"github.com/google/uuid"
	"github.com/raystack/salt/log"
	"github.com/robfig/cron/v3"
	"google.golang.org/grpc/metadata"
)

var (
	ErrNoSession       = errors.New("no session")
	ErrDeletingSession = errors.New("error deleting session")
	refreshTime        = "0 0 * * *" // Once a day at midnight (UTC)
)

type Repository interface {
	Set(ctx context.Context, session *Session) error
	Get(ctx context.Context, id uuid.UUID) (*Session, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteExpiredSessions(ctx context.Context) error
	UpdateValidity(ctx context.Context, id uuid.UUID, validity time.Duration) error
	List(ctx context.Context, userID string) ([]*Session, error)
	UpdateSessionMetadata(ctx context.Context, id uuid.UUID, metadata SessionMetadata, updatedAt time.Time) error
}

type Service struct {
	repo     Repository
	validity time.Duration
	log      log.Logger
	cron     *cron.Cron
	Now      func() time.Time
}

func NewService(logger log.Logger, repo Repository, validity time.Duration) *Service {
	return &Service{
		log:      logger,
		repo:     repo,
		cron:     cron.New(),
		validity: validity,
		Now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (s Service) Create(ctx context.Context, userID string, metadata SessionMetadata) (*Session, error) {
	now := s.Now()

	sess := &Session{
		ID:              uuid.New(),
		UserID:          userID,
		AuthenticatedAt: now,
		ExpiresAt:       now.Add(s.validity),
		CreatedAt:       now,
		UpdatedAt:       now,
		DeletedAt:       nil,
		Metadata:        metadata,
	}
	err := s.repo.Set(ctx, sess)
	if err != nil {
		s.log.Warn("failed to create session", "err", err)
		return nil, err
	}
	return sess, nil
}

// Refresh extends validity of session
func (s Service) Refresh(ctx context.Context, sessionID uuid.UUID) error {
	return s.repo.UpdateValidity(ctx, sessionID, s.validity)
}

// Delete marks a session as deleted without removing it from the database
func (s Service) Delete(ctx context.Context, sessionID uuid.UUID) error {
	return s.repo.Delete(ctx, sessionID)
}

func (s Service) ExtractFromContext(ctx context.Context) (*Session, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrNoSession
	}

	sessionHeaders := md.Get(consts.SessionIDGatewayKey)
	if len(sessionHeaders) == 0 || len(sessionHeaders[0]) == 0 {
		return nil, ErrNoSession
	}

	sessionID, err := uuid.Parse(sessionHeaders[0])
	if err != nil {
		return nil, ErrNoSession
	}
	return s.repo.Get(ctx, sessionID)
}

// InitSessions Initiates CronJob to delete expired sessions from the database
func (s Service) InitSessions(ctx context.Context) error {
	_, err := s.cron.AddFunc(refreshTime, func() {
		if err := s.repo.DeleteExpiredSessions(ctx); err != nil {
			s.log.Warn("failed to delete expired sessions", "err", err)
		}
	})
	if err != nil {
		return err
	}
	s.cron.Start()
	return nil
}

func (s Service) Close() error {
	return s.cron.Stop().Err()
}

// ListSessions returns all active sessions for a user
func (s Service) ListSessions(ctx context.Context, userID string) ([]*Session, error) {
	// Fetch all sessions for the user from repository
	sessions, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Filter for active sessions only
	var activeSessions []*Session
	now := s.Now()
	for _, session := range sessions {
		if session.IsValid(now) {
			activeSessions = append(activeSessions, session)
		}
	}

	return activeSessions, nil
}

func (s Service) PingSession(ctx context.Context, sessionID uuid.UUID, metadata SessionMetadata) error {
	now := s.Now()

	return s.repo.UpdateSessionMetadata(ctx, sessionID, metadata, now)
}

// GetSession retrieves a session by its ID
func (s Service) GetSession(ctx context.Context, sessionID uuid.UUID) (*Session, error) {
	return s.repo.Get(ctx, sessionID)
}
