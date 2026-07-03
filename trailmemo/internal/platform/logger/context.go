package logger

import (
	"context"

	"go.uber.org/zap"
)

type contextKey struct{}

// WithLogger 把带 request_id/user_id 等字段的 logger 放进 context。
func WithLogger(ctx context.Context, l *zap.Logger) context.Context {
	if l == nil {
		l = L()
	}
	return context.WithValue(ctx, contextKey{}, l)
}

// FromContext 从 context 取 logger；没有时退回全局 logger。
func FromContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return L()
	}
	if l, ok := ctx.Value(contextKey{}).(*zap.Logger); ok && l != nil {
		return l
	}
	return L()
}

// WithFields 给当前 context 中的 logger 追加字段，并返回新 context。
func WithFields(ctx context.Context, fields ...zap.Field) context.Context {
	return WithLogger(ctx, FromContext(ctx).With(fields...))
}
