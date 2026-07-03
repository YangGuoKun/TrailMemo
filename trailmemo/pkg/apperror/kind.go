package apperror

const (
	KindValidation = "validation" // 验证错误
	KindAuth       = "auth"       // 认证错误
	KindPermission = "permission" // 权限错误
	KindNotFound   = "not_found"  // 不存在错误
	KindConflict   = "conflict"   // 冲突错误
	KindDB         = "db"         // 数据库错误
	KindExternal   = "external"   // 外部错误
	KindInternal   = "internal"   // 内部错误
)
