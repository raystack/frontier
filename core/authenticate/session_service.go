package authenticate

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/internal/server/grpc_interceptors"
	"google.golang.org/grpc/metadata"
)

var (
	ErrNoSession = errors.New("no session")
)

type SessionRepository interface {
	Set(session *Session) error
	Get(id uuid.UUID) (*Session, error)
	Delete(id uuid.UUID) error
}

type SessionService struct {
	repo     SessionRepository
	validity time.Duration

	Now func() time.Time
}

func NewSessionManager(repo SessionRepository, validity time.Duration) *SessionService {
	return &SessionService{
		repo:     repo,
		validity: validity,
		Now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (s SessionService) Create(user user.User) (*Session, error) {
	sess := &Session{
		ID:              uuid.New(),
		UserID:          user.ID,
		AuthenticatedAt: s.Now(),
		ExpiresAt:       s.Now().Add(s.validity),
		CreatedAt:       s.Now(),
	}
	return sess, s.repo.Set(sess)
}

// Refresh extends validity of session
func (s SessionService) Refresh(session *Session) error {
	// TODO(kushsharma)
	panic("not implemented")
}

func (s SessionService) Delete(sessionID uuid.UUID) error {
	return s.repo.Delete(sessionID)
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
	return s.repo.Get(sessionID)
}
