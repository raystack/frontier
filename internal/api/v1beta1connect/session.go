package v1beta1connect

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/raystack/frontier/core/authenticate"
	frontiersession "github.com/raystack/frontier/core/authenticate/session"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SDK APIs
// Returns a list of all sessions for the current authenticated user.
func (h ConnectHandler) ListSessions(ctx context.Context, request *connect.Request[frontierv1beta1.ListSessionsRequest]) (*connect.Response[frontierv1beta1.ListSessionsResponse], error) {
	principal, err := h.authnService.GetPrincipal(ctx, authenticate.SessionClientAssertion)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	// Fetch all active sessions for the authenticated user
	sessions, err := h.sessionService.ListSessions(ctx, principal.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Transform domain sessions to protobuf sessions
	var pbSessions []*frontierv1beta1.Session
	for _, session := range sessions {
		pbSession, err := transformSessionToPB(session, principal.ID)
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
func transformSessionToPB(s *frontiersession.Session, currentUserID string) (*frontierv1beta1.Session, error) {
	metadata := &frontierv1beta1.Session_Meta{}
	if s.Metadata != nil {
		if os, ok := s.Metadata["operating_system"].(string); ok {
			metadata.OperatingSystem = os
		}
		if browser, ok := s.Metadata["browser"].(string); ok {
			metadata.Browser = browser
		}
		if ip, ok := s.Metadata["ip_address"].(string); ok {
			metadata.IpAddress = ip
		}
		if location, ok := s.Metadata["location"].(string); ok {
			metadata.Location = location
		}
	}

	return &frontierv1beta1.Session{
		Id:               s.ID.String(),
		Metadata:         metadata,
		IsCurrentSession: s.ID.String() == currentUserID,
		CreatedAt:        timestamppb.New(s.CreatedAt),
		UpdatedAt:        timestamppb.New(s.UpdatedAt),
	}, nil
}

// Revoke a specific session for the current authenticated user.
func (h ConnectHandler) RevokeSession(ctx context.Context, request *connect.Request[frontierv1beta1.RevokeSessionRequest]) (*connect.Response[frontierv1beta1.RevokeSessionResponse], error) {
	if _, err := h.authnService.GetPrincipal(ctx, authenticate.SessionClientAssertion); err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	if err := request.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	id, err := uuid.Parse(request.Msg.GetSessionId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if err := h.sessionService.SoftDelete(ctx, id, time.Now()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&frontierv1beta1.RevokeSessionResponse{}), nil
}

// Ping user current active session.
func (h ConnectHandler) PingUserSession(ctx context.Context, request *connect.Request[frontierv1beta1.PingUserSessionRequest]) (*connect.Response[frontierv1beta1.PingUserSessionResponse], error) {
	return nil, nil
}

// Admin APIs
// Returns a list of all sessions for a specific user.
func (h ConnectHandler) ListUserSessions(ctx context.Context, request *connect.Request[frontierv1beta1.ListUserSessionsRequest]) (*connect.Response[frontierv1beta1.ListUserSessionsResponse], error) {
	return nil, nil
}

// Revoke a specific session for a specific user (admin only).
func (h ConnectHandler) RevokeUserSession(ctx context.Context, request *connect.Request[frontierv1beta1.RevokeUserSessionRequest]) (*connect.Response[frontierv1beta1.RevokeUserSessionResponse], error) {
	if err := request.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	sessionID, err := uuid.Parse(request.Msg.GetSessionId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if err := h.sessionService.SoftDelete(ctx, sessionID, time.Now()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&frontierv1beta1.RevokeUserSessionResponse{}), nil
}
