package session

import (
	"context"
	"errors"
	"time"

	"github.com/odpf/shield/internal/server/consts"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

var (
	ErrNoSession       = errors.New("no session")
	ErrDeletingSession = errors.New("error deleting session")
)

type Repository interface {
	Set(ctx context.Context, session *Session) error
	Get(ctx context.Context, id uuid.UUID) (*Session, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Service struct {
	repo     Repository
	validity time.Duration

	Now func() time.Time
}

func NewService(repo Repository, validity time.Duration) *Service {
	return &Service{
		repo:     repo,
		validity: validity,
		Now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (s Service) Create(ctx context.Context, userID string) (*Session, error) {
	sess := &Session{
		ID:              uuid.New(),
		UserID:          userID,
		AuthenticatedAt: s.Now(),
		ExpiresAt:       s.Now().Add(s.validity),
		CreatedAt:       s.Now(),
	}
	return sess, s.repo.Set(ctx, sess)
}

// Refresh extends validity of session
func (s Service) Refresh(session *Session) error {
	// TODO(kushsharma)
	panic("not implemented")
}

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
