package v1beta1connect

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	frontierlogger "github.com/raystack/frontier/pkg/logger"
	"github.com/raystack/frontier/pkg/server/consts"
)

// ErrorLogger provides centralized error logging functionality for Connect handlers
type ErrorLogger struct{}

// NewErrorLogger creates a new ErrorLogger instance
func NewErrorLogger() *ErrorLogger {
	return &ErrorLogger{}
}

// LogServiceError logs detailed service error information with context
func (e *ErrorLogger) LogServiceError(ctx context.Context, req connect.AnyRequest, operation string, err error, contextArgs ...any) {
	logger := frontierlogger.FromContext(ctx)
	args := []any{
		"operation", operation,
		"request_id", e.getRequestID(req),
		"error_type", fmt.Sprintf("%T", err),
		"error", err,
	}
	args = append(args, contextArgs...)
	logger.Error(fmt.Sprintf("%s operation failed", operation), args...)
}

// LogUnexpectedError logs additional context for unexpected internal errors
func (e *ErrorLogger) LogUnexpectedError(ctx context.Context, req connect.AnyRequest, operation string, err error, contextArgs ...any) {
	logger := frontierlogger.FromContext(ctx)
	args := []any{
		"operation", operation,
		"request_id", e.getRequestID(req),
		"error_chain", fmt.Sprintf("%+v", err),
		"error", err,
	}
	args = append(args, contextArgs...)
	logger.Error(fmt.Sprintf("unexpected error in %s", operation), args...)
}

// LogTransformError logs protobuf transformation errors
func (e *ErrorLogger) LogTransformError(ctx context.Context, req connect.AnyRequest, operation string, entityID string, err error) {
	logger := frontierlogger.FromContext(ctx)
	logger.Error("protobuf transformation failed",
		"operation", operation,
		"request_id", e.getRequestID(req),
		"entity_id", entityID,
		"error_type", fmt.Sprintf("%T", err),
		"error", err)
}

// getRequestID extracts request ID from Connect request
func (e *ErrorLogger) getRequestID(req connect.AnyRequest) string {
	return req.Header().Get(consts.RequestIDHeader)
}
