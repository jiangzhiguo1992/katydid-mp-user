package log

import (
	"context"
	"go.uber.org/zap"
)

type contextKey int

const (
	requestIDKey contextKey = iota
	userIDKey
)

// ContextWithRequestID 添加请求ID到上下文
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// ContextWithUserID 添加用户ID到上下文
func ContextWithUserID(ctx context.Context, userID interface{}) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// extractContextFields 提取上下文字段
func extractContextFields(ctx context.Context) []zap.Field {
	fields := make([]zap.Field, 0, 2)

	if reqID, ok := ctx.Value(requestIDKey).(string); ok {
		fields = append(fields, zap.String("request_id", reqID))
	}

	if userID := ctx.Value(userIDKey); userID != nil {
		fields = append(fields, zap.Any("user_id", userID))
	}

	return fields
}