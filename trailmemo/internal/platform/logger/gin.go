package logger

import (
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const requestIDHeader = "X-Request-ID"

// RequestID 负责生成/透传 request_id，并把请求级 logger 注入 context。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader(requestIDHeader))
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Set(RequestIDKey, requestID)
		c.Header(requestIDHeader, requestID)

		reqLogger := L().With(
			zap.String("request_id", requestID),
			zap.String("trace_id", c.GetHeader("traceparent")),
		)
		c.Request = c.Request.WithContext(WithLogger(c.Request.Context(), reqLogger))
		c.Next()
	}
}

// AccessLog 在请求结束时输出一条访问日志，作为排查问题的入口。
func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		// 跳过一些路径，如 /healthz, /metrics, /debug 等
		if shouldSkipPath(c.FullPath(), c.Request.URL.Path) {
			return
		}

		status := c.Writer.Status()  // 获取响应状态码
		latency := time.Since(start) // 计算请求耗时
		fields := []zap.Field{       // 日志字段
			zap.String("event", "http_request_completed"),   // 事件名称
			zap.String("method", c.Request.Method),          // 请求方法
			zap.String("route", routePath(c)),               // 路由路径
			zap.String("path", c.Request.URL.Path),          // 请求路径
			zap.String("query", c.Request.URL.RawQuery),     // 查询参数
			zap.Int("status", status),                       // 响应状态码
			zap.Int64("latency_ms", latency.Milliseconds()), // 请求耗时（毫秒）
			zap.String("client_ip", c.ClientIP()),           // 客户端 IP
			zap.String("user_agent", c.Request.UserAgent()), // 用户代理
		}
		if userID := c.GetString(UserIDKey); userID != "" {
			fields = append(fields, zap.String("user_id", userID)) // 用户 ID
		}

		log := FromGinContext(c) // 从 gin.Context 中取 logger
		// 根据响应状态码判断日志级别
		switch {
		case status >= 500:
			log.Error("http_request_completed", fields...) // 错误日志
		case status >= 400:
			log.Warn("http_request_completed", fields...) // 警告日志
		default:
			log.Info("http_request_completed", fields...) // 信息日志
		}
	}
}

// Recovery 捕获 panic，并输出结构化错误日志，避免 gin.Recovery 的非结构化输出。
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() { // 捕获 panic 并输出日志
			if recovered := recover(); recovered != nil {
				FromGinContext(c).Error("panic_recovered",
					zap.String("event", "panic_recovered"),
					zap.String("error_kind", "panic"),
					zap.Any("error", recovered),
					zap.String("method", c.Request.Method),
					zap.String("route", routePath(c)),
					zap.String("path", c.Request.URL.Path),
					zap.String("stack", string(debug.Stack())),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":       http.StatusInternalServerError,
					"error_code": "INTERNAL_ERROR",
					"message":    "服务内部错误",
					"request_id": RequestIDFromGin(c),
				})
			}
		}()
		c.Next() // 继续处理请求
	}
}

// FromGinContext 从 gin.Context 对应的 request context 中取 logger。
func FromGinContext(c *gin.Context) *zap.Logger {
	if c == nil || c.Request == nil {
		return L()
	}
	return FromContext(c.Request.Context()) // 从 request context 中取 logger
}

// WithGinFields 给 gin 请求上下文中的 logger 追加字段，例如认证成功后追加 user_id。
func WithGinFields(c *gin.Context, fields ...zap.Field) {
	if c == nil || c.Request == nil {
		return
	}
	c.Request = c.Request.WithContext(WithFields(c.Request.Context(), fields...))
}

// RequestIDFromGin 从 gin.Context 中取 request_id。
func RequestIDFromGin(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if requestID := c.GetString(RequestIDKey); requestID != "" {
		return requestID
	}
	return c.Writer.Header().Get(requestIDHeader)
}

// shouldSkipPath 判断是否应该跳过该路径。
func shouldSkipPath(route, path string) bool {
	cfg := ConfigValue()
	for _, skipPath := range cfg.Request.SkipPaths {
		if skipPath == "" {
			continue
		}
		if strings.HasPrefix(route, skipPath) || strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// routePath 从 gin.Context 中取路由路径。
func routePath(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if fullPath := c.FullPath(); fullPath != "" {
		return fullPath
	}
	return c.Request.URL.Path
}

// userIDUint64 从 gin.Context 中取用户 ID。
func userIDUint64(c *gin.Context) uint64 {
	if c == nil {
		return 0
	}
	value := c.GetString(UserIDKey)
	if value == "" {
		return 0
	}
	userID, _ := strconv.ParseUint(value, 10, 64)
	return userID
}
