package authenticate

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/odpf/salt/log"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/internal/server/grpc_interceptors"
	"github.com/robfig/cron/v3"
	"google.golang.org/grpc/metadata"
)

var (
	ErrNoSession       = errors.New("no session")
	ErrDeletingSession = errors.New("error deleting session")
	refreshTime        = "0 3 * * *" // Once a day at midnight (UTC)
)

type SessionRepository interface {
	Set(ctx context.Context, session *Session) error
	Get(ctx context.Context, id uuid.UUID) (*Session, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteExpiredSessions(ctx context.Context, logger log.Logger) error
}

type SessionService struct {
	repo     SessionRepository
	validity time.Duration
	log      log.Logger
	Now      func() time.Time
}

func NewSessionManager(repo SessionRepository, validity time.Duration, logger log.Logger) *SessionService {
	return &SessionService{
		log:      logger,
		repo:     repo,
		validity: validity,
		Now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (s SessionService) Create(ctx context.Context, user user.User) (*Session, error) {
	sess := &Session{
		ID:              uuid.New(),
		UserID:          user.ID,
		AuthenticatedAt: s.Now(),
		ExpiresAt:       s.Now().Add(s.validity),
		CreatedAt:       s.Now(),
	}
	return sess, s.repo.Set(ctx, sess)
}

// Refresh extends validity of session
func (s SessionService) Refresh(session *Session) error {
	// TODO(kushsharma)
	panic("not implemented")
}

func (s SessionService) Delete(ctx context.Context, sessionID uuid.UUID) error {
	return s.repo.Delete(ctx, sessionID)
}

func (s SessionService) ExtractFromMD(ctx context.Context) (*Session, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrNoSession
	}

	sessionHeaders := md.Get(grpc_interceptors.SessionIDGatewayKey)
	if len(sessionHeaders) == 0 || len(sessionHeaders[0]) == 0 {
		return nil, ErrNoSession
	}

	sessionID, err := uuid.Parse(sessionHeaders[0])
	if err != nil {
		return nil, ErrNoSession
	}
	return s.repo.Get(ctx, sessionID)
}

// Initiates CronJob to delete expired sessions
func (s SessionService) RemoveExpiredSessions(ctx context.Context) error {
	cron := cron.New()
	_, err := cron.AddFunc(refreshTime, func() {
		if err := s.repo.DeleteExpiredSessions(ctx, s.log); err != nil {
			s.log.Warn("failed to delete expired sessions", "err", err)
		}
	})
	if err != nil {
		return err
	}
	cron.Start()
	// how to stop cron?
	// if a defer and select{} statement is used, the program will never exit this function
	return nil
}
