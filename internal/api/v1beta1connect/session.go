package v1beta1connect

import (
	"context"
	"errors"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/raystack/frontier/core/authenticate"
	frontiersession "github.com/raystack/frontier/core/authenticate/session"
	frontiererrors "github.com/raystack/frontier/pkg/errors"
	sessionutils "github.com/raystack/frontier/pkg/session"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SDK APIs
// Returns a list of all sessions for the current authenticated user.
func (h *ConnectHandler) ListSessions(ctx context.Context, request *connect.Request[frontierv1beta1.ListSessionsRequest]) (*connect.Response[frontierv1beta1.ListSessionsResponse], error) {
	principal, err := h.authnService.GetPrincipal(ctx, authenticate.SessionClientAssertion)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	var currentSessionID string
	if currentSession, err := h.sessionService.ExtractFromContext(ctx); err == nil {
		currentSessionID = currentSession.ID.String()
	}

	// Fetch all active sessions for the authenticated user
	sessions, err := h.sessionService.List(ctx, principal.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Transform domain sessions to protobuf sessions
	var pbSessions []*frontierv1beta1.Session
	for _, session := range sessions {
		pbSession, err := transformSessionToPB(session, currentSessionID)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		pbSessions = append(pbSessions, pbSession)
	}

	return connect.NewResponse(&frontierv1beta1.ListSessionsResponse{
		Sessions: pbSessions,
	}), nil
}

// transformSessionToPB converts a domain Session to a protobuf
func transformSessionToPB(s *frontiersession.Session, currentSessionID string) (*frontierv1beta1.Session, error) {
	city, country := strings.TrimSpace(s.Metadata.Location.City), strings.TrimSpace(s.Metadata.Location.Country)

	metadata := &frontierv1beta1.Session_Meta{
		OperatingSystem: s.Metadata.OperatingSystem,
		Browser:         s.Metadata.Browser,
		IpAddress:       s.Metadata.IpAddress,
		Location: func() string {
			if city == "" && country == "" {
				return ""
			}
			if city != "" && country != "" {
				return city + ", " + country
			}
			return city + country
		}(),
	}

	return &frontierv1beta1.Session{
		Id:               s.ID.String(),
		Metadata:         metadata,
		IsCurrentSession: s.ID.String() == currentSessionID,
		CreatedAt:        timestamppb.New(s.CreatedAt),
		UpdatedAt:        timestamppb.New(s.UpdatedAt),
	}, nil
}

// Revoke a specific session for the current authenticated user.
func (h *ConnectHandler) RevokeSession(ctx context.Context, request *connect.Request[frontierv1beta1.RevokeSessionRequest]) (*connect.Response[frontierv1beta1.RevokeSessionResponse], error) {
	principal, err := h.authnService.GetPrincipal(ctx, authenticate.SessionClientAssertion)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	if err := request.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	sessionID, err := uuid.Parse(request.Msg.GetSessionId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	session, err := h.sessionService.GetByID(ctx, sessionID)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	if session.UserID != principal.ID {
		return nil, connect.NewError(connect.CodeNotFound, frontiererrors.ErrNotFound)
	}

	if err := h.sessionService.Delete(ctx, sessionID); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&frontierv1beta1.RevokeSessionResponse{}), nil
}

// Ping user current active session.
func (h *ConnectHandler) PingUserSession(ctx context.Context, request *connect.Request[frontierv1beta1.PingUserSessionRequest]) (*connect.Response[frontierv1beta1.PingUserSessionResponse], error) {
	session, err := h.sessionService.ExtractFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	if !session.IsValid(time.Now().UTC()) {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	sessionMetadata := sessionutils.ExtractSessionMetadata(ctx, request, h.authConfig.Session.Headers)

	if err := h.sessionService.Ping(ctx, session.ID, sessionMetadata); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&frontierv1beta1.PingUserSessionResponse{}), nil
}

// Admin APIs
// Returns a list of all sessions for a specific user.
func (h *ConnectHandler) ListUserSessions(ctx context.Context, request *connect.Request[frontierv1beta1.ListUserSessionsRequest]) (*connect.Response[frontierv1beta1.ListUserSessionsResponse], error) {
	if err := request.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Manual validation for user_id since protobuf validation is not working
	userID := request.Msg.GetUserId()
	if userID == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("user_id is required"))
	}

	// Validate UUID format
	if _, err := uuid.Parse(userID); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid user_id format: must be a valid UUID"))
	}

	sessions, err := h.sessionService.List(ctx, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	var pbSessions []*frontierv1beta1.Session
	for _, session := range sessions {
		pbSession, err := transformSessionToPB(session, "")
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		pbSessions = append(pbSessions, pbSession)
	}

	return connect.NewResponse(&frontierv1beta1.ListUserSessionsResponse{
		Sessions: pbSessions,
	}), nil
}

// Revoke a specific session for a specific user (admin only).
func (h *ConnectHandler) RevokeUserSession(ctx context.Context, request *connect.Request[frontierv1beta1.RevokeUserSessionRequest]) (*connect.Response[frontierv1beta1.RevokeUserSessionResponse], error) {
	if err := request.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	sessionID, err := uuid.Parse(request.Msg.GetSessionId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if err := h.sessionService.Delete(ctx, sessionID); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&frontierv1beta1.RevokeUserSessionResponse{}), nil
}
