package consts

import "context"

type sessionMetadataKey string

const (
	SessionMetadataKey = sessionMetadataKey("session_metadata")
)

func WithSessionMetadata(ctx context.Context, metadata map[string]any) context.Context {
	return context.WithValue(ctx, SessionMetadataKey, metadata)
}

func GetSessionMetadata(ctx context.Context) (map[string]any, bool) {
	metadata, ok := ctx.Value(SessionMetadataKey).(map[string]any)
	return metadata, ok
}
