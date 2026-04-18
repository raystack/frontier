package v1beta1connect

import (
	"context"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
)

// ErrorLogger provides centralized error logging functionality for Connect handlers
type ErrorLogger struct{}

// NewErrorLogger creates a new ErrorLogger instance
func NewErrorLogger() *ErrorLogger {
	return &ErrorLogger{}
}

// LogServiceError logs detailed service error information with context
func (e *ErrorLogger) LogServiceError(ctx context.Context, req connect.AnyRequest, operation string, err error, contextArgs ...any) {
	args := []any{
		"operation", operation,
		"error_type", fmt.Sprintf("%T", err),
		"error", err,
	}
	args = append(args, contextArgs...)
	slog.ErrorContext(ctx, fmt.Sprintf("%s operation failed", operation), args...)
}

// LogUnexpectedError logs additional context for unexpected internal errors
func (e *ErrorLogger) LogUnexpectedError(ctx context.Context, req connect.AnyRequest, operation string, err error, contextArgs ...any) {
	args := []any{
		"operation", operation,
		"error_chain", fmt.Sprintf("%+v", err),
		"error", err,
	}
	args = append(args, contextArgs...)
	slog.ErrorContext(ctx, fmt.Sprintf("unexpected error in %s", operation), args...)
}

// LogTransformError logs protobuf transformation errors
func (e *ErrorLogger) LogTransformError(ctx context.Context, req connect.AnyRequest, operation string, entityID string, err error) {
	slog.ErrorContext(ctx, "protobuf transformation failed",
		"operation", operation,
		"entity_id", entityID,
		"error_type", fmt.Sprintf("%T", err),
		"error", err)
}
