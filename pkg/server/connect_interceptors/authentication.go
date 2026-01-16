package connectinterceptors

import (
	"context"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/auditrecord"
	"github.com/raystack/frontier/core/authenticate"
	frontiersession "github.com/raystack/frontier/core/authenticate/session"
	"github.com/raystack/frontier/internal/api/v1beta1connect"
	sessionutils "github.com/raystack/frontier/pkg/session"
)

type AuthenticationInterceptor struct {
	h                   *v1beta1connect.ConnectHandler
	sessionHeaderConfig authenticate.SessionMetadataHeaders
}

func NewAuthenticationInterceptor(h *v1beta1connect.ConnectHandler, sessionHeaderConfig authenticate.SessionMetadataHeaders) *AuthenticationInterceptor {
	return &AuthenticationInterceptor{
		h:                   h,
		sessionHeaderConfig: sessionHeaderConfig,
	}
}

func (i *AuthenticationInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if authenticationSkipList[req.Spec().Procedure] {
			return next(ctx, req)
		}

		principal, err := i.h.GetLoggedInPrincipal(ctx)
		if err != nil {
			return nil, err
		}
		ctx = authenticate.SetContextWithPrincipal(ctx, &principal)
		ctx = audit.SetContextWithActor(ctx, audit.Actor{
			ID:   principal.ID,
			Type: principal.Type,
		})

		isSuperUser := false
		err = i.h.IsSuperUser(ctx, req)
		if err == nil {
			isSuperUser = true
		}
		ctx = authenticate.SetSuperUserInContext(ctx, isSuperUser)

		sessionMetadata := sessionutils.ExtractSessionMetadata(ctx, req, i.sessionHeaderConfig)
		ctx = frontiersession.SetSessionMetadataInContext(ctx, sessionMetadata)

		// Set audit record actor context - for repositories and audit consumers
		actorName, actorTitle := authenticate.GetPrincipalNameAndTitle(&principal)
		ctx = auditrecord.SetAuditRecordActorContext(ctx, auditrecord.Actor{
			ID:       principal.ID,
			Type:     principal.Type,
			Name:     actorName,
			Title:    actorTitle,
			Metadata: nil,
		})
		return next(ctx, req)
	})
}

func (i *AuthenticationInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return connect.StreamingClientFunc(func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		conn := next(ctx, spec)
		return conn
	})
}

func (i *AuthenticationInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		if authenticationSkipList[conn.Spec().Procedure] {
			return next(ctx, conn)
		}

		principal, err := i.h.GetLoggedInPrincipal(ctx)
		if err != nil {
			return err
		}
		ctx = authenticate.SetContextWithPrincipal(ctx, &principal)
		ctx = audit.SetContextWithActor(ctx, audit.Actor{
			ID:   principal.ID,
			Type: principal.Type,
		})
		return next(ctx, conn)
	})
}

// authenticationSkipList stores path to skip authentication, by default its enabled for all requests
var authenticationSkipList = map[string]bool{
	"/raystack.frontier.v1beta1.FrontierService/ListAuthStrategies":     true,
	"/raystack.frontier.v1beta1.FrontierService/Authenticate":           true,
	"/raystack.frontier.v1beta1.FrontierService/AuthCallback":           true,
	"/raystack.frontier.v1beta1.FrontierService/ListMetaSchemas":        true,
	"/raystack.frontier.v1beta1.FrontierService/GetMetaSchema":          true,
	"/raystack.frontier.v1beta1.FrontierService/BillingWebhookCallback": true,
}
