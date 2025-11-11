package v1beta1connect

import (
	"context"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/raystack/frontier/core/authenticate"
	frontiersession "github.com/raystack/frontier/core/authenticate/session"
	sessionutils "github.com/raystack/frontier/pkg/session"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SDK APIs
// Returns a list of all sessions for the current authenticated user.
func (h *ConnectHandler) ListSessions(ctx context.Context, request *connect.Request[frontierv1beta1.ListSessionsRequest]) (*connect.Response[frontierv1beta1.ListSessionsResponse], error) {
	errorLogger := NewErrorLogger()

	principal, err := h.authnService.GetPrincipal(ctx, authenticate.SessionClientAssertion)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "ListSessions.GetPrincipal", err)
		return nil, connect.NewError(connect.CodeUnauthenticated, ErrUnauthenticated)
	}

	var currentSessionID string
	if currentSession, err := h.sessionService.ExtractFromContext(ctx); err == nil {
		currentSessionID = currentSession.ID.String()
	}

	// Fetch all active sessions for the authenticated user
	sessions, err := h.sessionService.List(ctx, principal.ID)
	if err != nil {
		errorLogger.LogUnexpectedError(ctx, request, "ListSessions.List", err,
			zap.String("user_id", principal.ID))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	// Transform domain sessions to protobuf sessions
	var pbSessions []*frontierv1beta1.Session
	for _, session := range sessions {
		pbSession, err := transformSessionToPB(session, currentSessionID)
		if err != nil {
			errorLogger.LogTransformError(ctx, request, "ListSessions", session.ID.String(), err)
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		pbSessions = append(pbSessions, pbSession)
	}

	return connect.NewResponse(&frontierv1beta1.ListSessionsResponse{
		Sessions: pbSessions,
	}), nil
}

// transformSessionMetadataToPB converts session metadata to protobuf format
func transformSessionMetadataToPB(metadata frontiersession.SessionMetadata) *frontierv1beta1.Session_Meta {
	city, country := strings.TrimSpace(metadata.Location.City), strings.TrimSpace(metadata.Location.Country)
	latitude, longitude := strings.TrimSpace(metadata.Location.Latitude), strings.TrimSpace(metadata.Location.Longitude)

	return &frontierv1beta1.Session_Meta{
		OperatingSystem: metadata.OperatingSystem,
		Browser:         metadata.Browser,
		IpAddress:       metadata.IpAddress,
		Location: &frontierv1beta1.Session_Meta_Location{
			City:      city,
			Country:   country,
			Latitude:  latitude,
			Longitude: longitude,
		},
	}
}

// transformSessionToPB converts a domain Session to a protobuf
func transformSessionToPB(s *frontiersession.Session, currentSessionID string) (*frontierv1beta1.Session, error) {
	metadata := transformSessionMetadataToPB(s.Metadata)

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
	errorLogger := NewErrorLogger()

	principal, err := h.authnService.GetPrincipal(ctx, authenticate.SessionClientAssertion)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "RevokeSession.GetPrincipal", err)
		return nil, connect.NewError(connect.CodeUnauthenticated, ErrUnauthenticated)
	}

	if err := request.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	sessionID, err := uuid.Parse(request.Msg.GetSessionId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrInvalidSessionID)
	}

	session, err := h.sessionService.GetByID(ctx, sessionID)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "RevokeSession.GetByID", err,
			zap.String("session_id", sessionID.String()))
		return nil, connect.NewError(connect.CodeNotFound, ErrSessionNotFound)
	}

	if session.UserID != principal.ID {
		return nil, connect.NewError(connect.CodeNotFound, ErrSessionNotFound)
	}

	if err := h.sessionService.Delete(ctx, sessionID); err != nil {
		errorLogger.LogUnexpectedError(ctx, request, "RevokeSession", err,
			zap.String("session_id", sessionID.String()),
			zap.String("user_id", principal.ID))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.RevokeSessionResponse{}), nil
}

// Ping user current active session.
func (h *ConnectHandler) PingUserSession(ctx context.Context, request *connect.Request[frontierv1beta1.PingUserSessionRequest]) (*connect.Response[frontierv1beta1.PingUserSessionResponse], error) {
	errorLogger := NewErrorLogger()

	session, err := h.sessionService.ExtractFromContext(ctx)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "PingUserSession.ExtractFromContext", err)
		return nil, connect.NewError(connect.CodeUnauthenticated, ErrUnauthenticated)
	}

	if !session.IsValid(time.Now().UTC()) {
		return nil, connect.NewError(connect.CodeUnauthenticated, ErrUnauthenticated)
	}

	sessionMetadata := sessionutils.ExtractSessionMetadata(ctx, request, h.authConfig.Session.Headers)

	if err := h.sessionService.Ping(ctx, session.ID, sessionMetadata); err != nil {
		errorLogger.LogUnexpectedError(ctx, request, "PingUserSession", err,
			zap.String("session_id", session.ID.String()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	// Fetch updated session to get latest metadata
	updatedSession, err := h.sessionService.GetByID(ctx, session.ID)
	if err != nil {
		errorLogger.LogUnexpectedError(ctx, request, "PingUserSession.GetByID", err,
			zap.String("session_id", session.ID.String()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	// Convert session metadata to proto format
	metadata := transformSessionMetadataToPB(updatedSession.Metadata)

	return connect.NewResponse(&frontierv1beta1.PingUserSessionResponse{
		Metadata: metadata,
	}), nil
}

// Admin APIs
// Returns a list of all sessions for a specific user.
func (h *ConnectHandler) ListUserSessions(ctx context.Context, request *connect.Request[frontierv1beta1.ListUserSessionsRequest]) (*connect.Response[frontierv1beta1.ListUserSessionsResponse], error) {
	errorLogger := NewErrorLogger()

	if err := request.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	// Manual validation for user_id since protobuf validation is not working
	userID := request.Msg.GetUserId()
	if userID == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrInvalidUserID)
	}

	// Validate UUID format
	if _, err := uuid.Parse(userID); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrInvalidUserID)
	}

	sessions, err := h.sessionService.List(ctx, userID)
	if err != nil {
		errorLogger.LogUnexpectedError(ctx, request, "ListUserSessions", err,
			zap.String("user_id", userID))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	var pbSessions []*frontierv1beta1.Session
	for _, session := range sessions {
		pbSession, err := transformSessionToPB(session, "")
		if err != nil {
			errorLogger.LogTransformError(ctx, request, "ListUserSessions", session.ID.String(), err)
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		pbSessions = append(pbSessions, pbSession)
	}

	return connect.NewResponse(&frontierv1beta1.ListUserSessionsResponse{
		Sessions: pbSessions,
	}), nil
}

// Revoke a specific session for a specific user (admin only).
func (h *ConnectHandler) RevokeUserSession(ctx context.Context, request *connect.Request[frontierv1beta1.RevokeUserSessionRequest]) (*connect.Response[frontierv1beta1.RevokeUserSessionResponse], error) {
	errorLogger := NewErrorLogger()

	if err := request.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	sessionID, err := uuid.Parse(request.Msg.GetSessionId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrInvalidSessionID)
	}

	if err := h.sessionService.Delete(ctx, sessionID); err != nil {
		errorLogger.LogUnexpectedError(ctx, request, "RevokeUserSession", err,
			zap.String("session_id", sessionID.String()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.RevokeUserSessionResponse{}), nil
}
