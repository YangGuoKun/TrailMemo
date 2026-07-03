package apperror

const (
	CodeInvalidParams    = "INVALID_PARAMS"     // 无效参数
	CodeUnauthorized     = "UNAUTHORIZED"       // 未授权
	CodeTokenExpired     = "TOKEN_EXPIRED"      // 令牌过期
	CodePermissionDenied = "PERMISSION_DENIED"  // 权限被拒绝
	CodeResourceNotFound = "RESOURCE_NOT_FOUND" // 资源不存在
	CodeConflict         = "CONFLICT"           // 冲突
	CodeDBError          = "DB_ERROR"           // 数据库错误
	CodeExternalError    = "EXTERNAL_ERROR"     // 外部错误
	CodeInternalError    = "INTERNAL_ERROR"     // 内部错误

	CodeUserNotFound        = "USER_NOT_FOUND"          // 用户不存在
	CodeUserAlreadyExists   = "USER_ALREADY_EXISTS"     // 用户已存在
	CodeUserUpdateFailed    = "USER_UPDATE_FAILED"      // 用户更新失败
	CodeInvalidCredentials  = "INVALID_CREDENTIALS"     // 无效的凭证
	CodeWechatLoginFailed   = "WECHAT_LOGIN_FAILED"     // 微信登录失败
	CodeRouteNotFound       = "ROUTE_NOT_FOUND"         // 路由不存在
	CodeRoutePermission     = "ROUTE_PERMISSION_DENIED" // 路由权限被拒绝
	CodeRouteCreateFailed   = "ROUTE_CREATE_FAILED"     // 路由创建失败
	CodeRouteUpdateFailed   = "ROUTE_UPDATE_FAILED"     // 路由更新失败
	CodeRouteDeleteFailed   = "ROUTE_DELETE_FAILED"     // 路由删除失败
	CodeRouteCopyFailed     = "ROUTE_COPY_FAILED"       // 路由复制失败
	CodeCheckpointNotFound  = "CHECKPOINT_NOT_FOUND"    // 检查点不存在
	CodeCheckinNotFound     = "CHECKIN_NOT_FOUND"       // 签到不存在
	CodeCheckinPermission   = "CHECKIN_PERMISSION_DENIED"
	CodeCheckinCreateFailed = "CHECKIN_CREATE_FAILED" // 签到创建失败
	CodeCheckinUpdateFailed = "CHECKIN_UPDATE_FAILED"
	CodeCheckinDeleteFailed = "CHECKIN_DELETE_FAILED"
	CodeCheckinConflict     = "CHECKIN_CONFLICT"
	CodePostNotFound        = "POST_NOT_FOUND"            // 帖子不存在
	CodePostPermission      = "POST_PERMISSION_DENIED"    // 帖子权限被拒绝
	CodePostCreateFailed    = "POST_CREATE_FAILED"        // 帖子创建失败
	CodePostUpdateFailed    = "POST_UPDATE_FAILED"        // 帖子更新失败
	CodePostDeleteFailed    = "POST_DELETE_FAILED"        // 帖子删除失败
	CodeCommentNotFound     = "COMMENT_NOT_FOUND"         // 评论不存在
	CodeCommentPermission   = "COMMENT_PERMISSION_DENIED" // 评论权限被拒绝
	CodeCommentCreateFailed = "COMMENT_CREATE_FAILED"     // 评论创建失败
	CodeCommentUpdateFailed = "COMMENT_UPDATE_FAILED"     // 评论更新失败
	CodeCommentDeleteFailed = "COMMENT_DELETE_FAILED"     // 评论删除失败
	CodeLikeConflict        = "LIKE_CONFLICT"             // 点赞冲突
	CodeFavoriteConflict    = "FAVORITE_CONFLICT"         // 收藏冲突
)
