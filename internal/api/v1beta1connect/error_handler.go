package v1beta1connect

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/frontier/pkg/server/consts"
	"go.uber.org/zap"
)

// ErrorLogger provides centralized error logging functionality for Connect handlers
type ErrorLogger struct{}

// NewErrorLogger creates a new ErrorLogger instance
func NewErrorLogger() *ErrorLogger {
	return &ErrorLogger{}
}

// LogServiceError logs detailed service error information with context
func (e *ErrorLogger) LogServiceError(ctx context.Context, req connect.AnyRequest, operation string, err error, contextFields ...zap.Field) {
	logger := grpczap.Extract(ctx)
	requestID := e.getRequestID(req)

	// Build base fields for logging
	baseFields := []zap.Field{
		zap.String("operation", operation),
		zap.String("request_id", requestID),
		zap.String("error_type", fmt.Sprintf("%T", err)),
		zap.Error(err),
	}
	baseFields = append(baseFields, contextFields...)

	// Log detailed error
	logger.Error(fmt.Sprintf("%s operation failed", operation), baseFields...)
}

// LogUnexpectedError logs additional context for unexpected internal errors
func (e *ErrorLogger) LogUnexpectedError(ctx context.Context, req connect.AnyRequest, operation string, err error, contextFields ...zap.Field) {
	logger := grpczap.Extract(ctx)
	requestID := e.getRequestID(req)

	// Build base fields for logging
	baseFields := []zap.Field{
		zap.String("operation", operation),
		zap.String("request_id", requestID),
		zap.String("error_chain", fmt.Sprintf("%+v", err)),
		zap.Error(err),
	}
	baseFields = append(baseFields, contextFields...)

	logger.Error(fmt.Sprintf("unexpected error in %s", operation), baseFields...)
}

// LogTransformError logs protobuf transformation errors
func (e *ErrorLogger) LogTransformError(ctx context.Context, req connect.AnyRequest, operation string, entityID string, err error) {
	logger := grpczap.Extract(ctx)
	requestID := e.getRequestID(req)

	logger.Error("protobuf transformation failed",
		zap.String("operation", operation),
		zap.String("request_id", requestID),
		zap.String("entity_id", entityID),
		zap.String("error_type", fmt.Sprintf("%T", err)),
		zap.Error(err))
}

// getRequestID extracts request ID from Connect request
func (e *ErrorLogger) getRequestID(req connect.AnyRequest) string {
	return req.Header().Get(consts.RequestIDHeader)
}

