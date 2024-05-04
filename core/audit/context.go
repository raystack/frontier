package audit

import (
	"context"

	"github.com/raystack/frontier/pkg/server/consts"
)

// GetService returns the audit service from the context if set
// otherwise it returns a new service with a write only repository
func GetService(ctx context.Context) *Service {
	u, ok := ctx.Value(consts.AuditServiceContextKey).(*Service)
	if !ok {
		return NewService("default", NewNoopRepository(), NewNoopWebhookService())
	}
	return u
}

func SetContextWithService(ctx context.Context, p *Service) context.Context {
	return context.WithValue(ctx, consts.AuditServiceContextKey, p)
}

func GetAuditor(ctx context.Context, orgID string) *Logger {
	return NewLogger(ctx, orgID)
}

func SetContextWithActor(ctx context.Context, actor Actor) context.Context {
	return context.WithValue(ctx, consts.AuditActorContextKey, actor)
}

func SetContextWithMetadata(ctx context.Context, md map[string]string) context.Context {
	existingMetadata, ok := ctx.Value(consts.AuditMetadataContextKey).(map[string]string)
	if !ok || existingMetadata == nil {
		return context.WithValue(ctx, consts.AuditMetadataContextKey, md)
	}

	// append new metadata
	for k, v := range md {
		existingMetadata[k] = v
	}
	return context.WithValue(ctx, consts.AuditMetadataContextKey, existingMetadata)
}

func defaultActorExtractor(ctx context.Context) (Actor, bool) {
	if actor, ok := ctx.Value(consts.AuditActorContextKey).(Actor); ok {
		return actor, true
	}
	return Actor{
		Name: "anonymous",
	}, false
}

func defaultMetadataExtractor(ctx context.Context) (map[string]string, bool) {
	if md, ok := ctx.Value(consts.AuditMetadataContextKey).(map[string]string); ok {
		return md, true
	}
	return map[string]string{}, false
}
