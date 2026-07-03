package logger

import (
	"context"

	"go.uber.org/zap"
)

// Audit 记录审计事件。第一阶段先输出到日志平台，后续可以在这里扩展为落库。
// ctx 用于获取带 request_id 的 logger，确保审计日志能关联到具体请求。
func Audit(ctx context.Context, action string, fields ...zap.Field) {
	if !activeConfig.Audit.Enabled {
		return
	}
	allFields := append([]zap.Field{
		zap.String("event", "audit_event"),
		zap.String("audit_action", action),
	}, fields...)
	FromContext(ctx).Info("audit_event", allFields...)
}
