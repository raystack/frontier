package v1beta1connect

import (
	"context"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/authenticate"
	frontiersession "github.com/raystack/frontier/core/authenticate/session"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SDK APIs
// Returns a list of all sessions for the current authenticated user.
func (h ConnectHandler) ListSessions(ctx context.Context, request *frontierv1beta1.ListSessionsRequest) (*frontierv1beta1.ListSessionsResponse, error) {
	principal, err := h.authnService.GetPrincipal(ctx, authenticate.SessionClientAssertion)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	// Fetch all active sessions for the authenticated user
	sessions, err := h.sessionService.ListSessions(ctx, principal.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Transform domain sessions to protobuf sessions
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

// transformSessionToPB converts a domain Session to a protobuf Session
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
func (h ConnectHandler) RevokeSession(ctx context.Context, request *frontierv1beta1.RevokeSessionRequest) (*frontierv1beta1.RevokeSessionResponse, error) {
	if _, err := h.authnService.GetPrincipal(ctx, authenticate.SessionClientAssertion); err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	if request.GetSessionId() == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	id, err := uuid.Parse(request.GetSessionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid session_id")
	}

	// TODO: instead of directly calling delete we need to mark it as deleted and delete after a day with a cron job.
	if err := h.sessionService.SoftDelete(ctx, id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &frontierv1beta1.RevokeSessionResponse{}, nil
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
