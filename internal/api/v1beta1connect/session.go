package v1beta1connect

import (
	"context"

	"github.com/raystack/frontier/core/authenticate"
	frontiersession "github.com/raystack/frontier/core/authenticate/session"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SDK APIs
// Returns a list of all sessions for the current authenticated user.
func (h ConnectHandler) ListSessions(ctx context.Context, request *frontierv1beta1.ListSessionsRequest) (*frontierv1beta1.ListSessionsResponse, error) {
	principal, err := h.authnService.GetPrincipal(ctx, authenticate.SessionClientAssertion)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	sessions, err := h.sessionService.ListSessions(ctx, principal.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var pbSessions []*frontierv1beta1.Session
    for _, session := range sessions {
        pbSession, err := transformSessionToPB(session, principal.ID)
        if err != nil {
            return nil, status.Error(codes.Internal, "error transforming session data")
        }
        pbSessions = append(pbSessions, pbSession)
    }

	return &frontierv1beta1.ListSessionsResponse{
		Sessions: pbSessions,
	}, nil
}

func transformSessionToPB(session *frontiersession.Session, currentUserID string) (*frontierv1beta1.Session, error) {
	// Check if this is the current session
	isCurrentSession := session.Id == currentUserID

	return &frontierv1beta1.Session{
		Id:               session.Id,
		Metadata:         &frontierv1beta1.Session_Meta{
			OperatingSystem: session.Metadata.OperatingSystem,
			Browser:         session.Metadata.Browser,
			IpAddress:       session.Metadata.IpAddress,
			Location:        session.Metadata.Location,
		},
		IsCurrentSession: isCurrentSession,
		CreatedAt:        session.CreatedAt,
		UpdatedAt:        session.UpdatedAt,
	}, nil
}

// Revoke a specific session for the current authenticated user.
func (h ConnectHandler) RevokeSession(ctx context.Context, request *frontierv1beta1.RevokeSessionRequest) (*frontierv1beta1.RevokeSessionResponse, error) {
	return nil, nil
}

// Ping user current active session.
func (h ConnectHandler) PingUserSession(ctx context.Context, request *frontierv1beta1.PingUserSessionRequest) (*frontierv1beta1.PingUserSessionResponse, error) {
	return nil, nil
}

// Admin APIs
// Returns a list of all sessions for a specific user.
func (h ConnectHandler) ListUserSessions(ctx context.Context, request *frontierv1beta1.ListUserSessionsRequest) (*frontierv1beta1.ListUserSessionsResponse, error) {
	return nil, nil
}

// Revoke a specific session for a specific user (admin only).
func (h ConnectHandler) RevokeUserSession(ctx context.Context, request *frontierv1beta1.RevokeUserSessionRequest) (*frontierv1beta1.RevokeUserSessionResponse, error) {
	return nil, nil
}
