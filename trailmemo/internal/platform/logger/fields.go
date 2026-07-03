package logger

// gin.Context key
const (
	RequestIDKey = "request_id"
	UserIDKey    = "user_id"
	TraceIDKey   = "trace_id"
)

// 事件名
const (
	EventAccessLog                = "http_request_completed"
	EventPanicRecovered           = "panic_recovered"
	EventAuthFailed               = "auth_failed"
	EventValidationFailed         = "validation_failed"
	EventPermissionDenied         = "permission_denied"
	EventServiceFailed            = "service_failed"
	EventGormError                = "gorm_error"
	EventGormSlowQuery            = "gorm_slow_query"
	EventGormQuery                = "gorm_query"
	EventTxBegin                  = "tx_begin"
	EventTxRollback               = "tx_rollback"
	EventTxCommitFailed           = "tx_commit_failed"
	EventExternalCallCompleted    = "external_call_completed"
	EventExternalCallFailed       = "external_call_failed"
	EventExternalCallSlow         = "external_call_slow"
	EventFallbackUsed             = "fallback_used"
	EventRetryExhausted           = "retry_exhausted"
	EventAuditEvent               = "audit_event"
	EventServerStarted            = "server_started"
	EventServerStopped            = "server_stopped"
	EventDatabaseInitFailed       = "database_init_failed"
	EventMigrationFailed          = "migration_failed"
	EventServerStartFailed        = "server_start_failed"
	EventRedisUnavailable         = "redis_unavailable"

	// 业务成功事件
	EventRouteCreated   = "route_created"
	EventRouteUpdated   = "route_updated"
	EventRouteDeleted   = "route_deleted"
	EventRouteCopied    = "route_copied"
	EventCheckinCreated = "checkin_created"
	EventPostCreated    = "post_created"
	EventCommentCreated = "comment_created"
	EventLikeToggled    = "like_toggled"
	EventFavoriteToggled = "favorite_toggled"
)

// 错误分类
const (
	KindValidation  = "validation"
	KindAuth        = "auth"
	KindPermission  = "permission"
	KindNotFound    = "not_found"
	KindConflict    = "conflict"
	KindDB          = "db"
	KindExternal    = "external"
	KindInternal    = "internal"
	KindPanic       = "panic"
)

// JSON 字段名
const (
	FieldEvent         = "event"
	FieldRequestID     = "request_id"
	FieldUserID        = "user_id"
	FieldModule        = "module"
	FieldOperation     = "operation"
	FieldEntityType    = "entity_type"
	FieldEntityID      = "entity_id"
	FieldErrorCode     = "error_code"
	FieldErrorKind     = "error_kind"
	FieldPublicMessage = "public_message"
	FieldResult        = "result"
	FieldAuditAction   = "audit_action"
	FieldReason        = "reason"
	FieldLatencyMs     = "latency_ms"
	FieldClientIP      = "client_ip"
	FieldUserAgent     = "user_agent"
	FieldMethod        = "method"
	FieldRoute         = "route"
	FieldPath          = "path"
	FieldQuery         = "query"
	FieldStatus        = "status"
	FieldSQL           = "sql"
	FieldRows          = "rows"
	FieldStack         = "stack"
)
