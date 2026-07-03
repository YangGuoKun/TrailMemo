package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/trailmemo/internal/middleware"
	platformlogger "github.com/trailmemo/internal/platform/logger"
	"github.com/trailmemo/internal/service"
	"github.com/trailmemo/pkg/apperror"
	"github.com/trailmemo/pkg/response"
)

type UserHandler struct {
	service service.UserService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		service: service.NewUserService(),
	}
}

func (h *UserHandler) RegisterRoutes(r *gin.RouterGroup) {
	userGroup := r.Group("/users")
	{
		userGroup.POST("/register", h.Register)
		userGroup.POST("/login", h.Login)
		userGroup.POST("/login/wechat", h.WechatLogin)
	}

	authGroup := r.Group("/users")
	authGroup.Use(middleware.JWTAuth())
	{
		authGroup.GET("/me", h.GetUserInfo)
		authGroup.PUT("/me", h.UpdateUserInfo)
		authGroup.PUT("/me/password", h.ChangePassword)
	}
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64" example:"testuser"`
	Password string `json:"password" binding:"required,min=6,max=128" example:"123456"`
	Phone    string `json:"phone" binding:"omitempty,e164" example:"+8613800138000"`
	Email    string `json:"email" binding:"omitempty,email" example:"test@example.com"`
}

// Register 用户注册
// @Summary 用户注册
// @Description 创建新用户账号
// @Tags 用户
// @Accept json
// @Produce json
// @Param body body RegisterRequest true "注册信息"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/v1/users/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "user"),
		zap.String("operation", "register"),
	)

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("validation_failed",
			zap.String("event", "validation_failed"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeInvalidParams, apperror.KindValidation, "参数错误", http.StatusBadRequest))
		return
	}

	log = log.With(zap.String("username", req.Username))

	user, err := h.service.Register(ctx, req.Username, req.Password, req.Phone, req.Email)
	if err != nil {
		log.Error("service_failed",
			zap.String("event", "service_failed"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeInvalidParams, apperror.KindValidation, err.Error(), http.StatusBadRequest))
		return
	}

	log.Info("user_registered",
		zap.String("event", "user_registered"),
		zap.String("entity_type", "user"),
		zap.Uint64("entity_id", user.ID),
	)
	platformlogger.Audit(c.Request.Context(), "user.register",
		zap.String("request_id", platformlogger.RequestIDFromGin(c)),
		zap.Uint64("user_id", user.ID),
		zap.String("entity_type", "user"),
		zap.Uint64("entity_id", user.ID),
		zap.String("result", "success"),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)

	response.Success(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
	})
}

type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"testuser"`
	Password string `json:"password" binding:"required" example:"123456"`
}

// Login 用户登录
// @Summary 用户登录
// @Description 使用用户名和密码登录
// @Tags 用户
// @Accept json
// @Produce json
// @Param body body LoginRequest true "登录信息"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/v1/users/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "user"),
		zap.String("operation", "login"),
	)

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("validation_failed",
			zap.String("event", "validation_failed"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeInvalidParams, apperror.KindValidation, "参数错误", http.StatusBadRequest))
		return
	}

	log = log.With(zap.String("username", req.Username))

	token, err := h.service.Login(ctx, req.Username, req.Password)
	if err != nil {
		log.Warn("login_failed",
			zap.String("event", "login_failed"),
			zap.String("error_code", apperror.CodeUnauthorized),
			zap.String("error_kind", apperror.KindAuth),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeUnauthorized, apperror.KindAuth, err.Error(), http.StatusUnauthorized))
		return
	}

	log.Info("user_login_success",
		zap.String("event", "user_login_success"),
	)
	platformlogger.Audit(c.Request.Context(), "user.login",
		zap.String("request_id", platformlogger.RequestIDFromGin(c)),
		zap.String("username", req.Username),
		zap.String("result", "success"),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)

	response.Success(c, gin.H{
		"token": token,
	})
}

type WechatLoginRequest struct {
	Code string `json:"code" binding:"required" example:"081aBc1234567"`
}

// WechatLogin 微信登录
// @Summary 微信小程序登录
// @Description 使用微信授权码登录
// @Tags 用户
// @Accept json
// @Produce json
// @Param body body WechatLoginRequest true "微信授权码"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/v1/users/login/wechat [post]
func (h *UserHandler) WechatLogin(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "user"),
		zap.String("operation", "wechat_login"),
	)

	var req WechatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("validation_failed",
			zap.String("event", "validation_failed"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeInvalidParams, apperror.KindValidation, "参数错误", http.StatusBadRequest))
		return
	}

	token, err := h.service.WechatLogin(ctx, req.Code)
	if err != nil {
		log.Warn("wechat_login_failed",
			zap.String("event", "wechat_login_failed"),
			zap.String("error_code", apperror.CodeUnauthorized),
			zap.String("error_kind", apperror.KindAuth),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeUnauthorized, apperror.KindAuth, err.Error(), http.StatusUnauthorized))
		return
	}

	log.Info("wechat_login_success",
		zap.String("event", "wechat_login_success"),
	)
	platformlogger.Audit(c.Request.Context(), "user.wechat_login",
		zap.String("request_id", platformlogger.RequestIDFromGin(c)),
		zap.String("result", "success"),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)

	response.Success(c, gin.H{
		"token": token,
	})
}

// GetUserInfo 获取当前用户信息
// @Summary 获取用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 用户
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/users/me [get]
func (h *UserHandler) GetUserInfo(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "user"),
		zap.String("operation", "get_user_info"),
	)

	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		log.Warn("unauthorized",
			zap.String("event", "unauthorized"),
			zap.String("error_code", apperror.CodeUnauthorized),
			zap.String("error_kind", apperror.KindAuth),
		)
		response.Unauthorized(c, "invalid or expired token")
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		log.Warn("invalid_user_id",
			zap.String("event", "invalid_user_id"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeInvalidParams, apperror.KindValidation, "用户ID无效", http.StatusBadRequest))
		return
	}

	log = log.With(zap.Uint64("user_id", userID))

	user, err := h.service.GetUserInfo(ctx, userID)
	if err != nil {
		log.Error("service_failed",
			zap.String("event", "service_failed"),
			zap.String("error_code", apperror.CodeDBError),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, err.Error(), http.StatusInternalServerError))
		return
	}

	if user == nil {
		log.Warn("user_not_found",
			zap.String("event", "user_not_found"),
		)
		response.NotFound(c, "user not found")
		return
	}

	response.Success(c, gin.H{
		"id":        user.ID,
		"username":  user.Username,
		"nickname":  user.Nickname,
		"avatar":    user.Avatar,
		"phone":     ptrToStr(user.Phone),
		"email":     ptrToStr(user.Email),
		"createdAt": user.CreatedAt,
	})
}

type UpdateUserInfoRequest struct {
	Nickname string `json:"nickname" example:"新昵称"`
	Avatar   string `json:"avatar" example:"https://example.com/avatar.jpg"`
}

// UpdateUserInfo 更新用户信息
// @Summary 更新用户信息
// @Description 更新当前登录用户的昵称和头像
// @Tags 用户
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body UpdateUserInfoRequest true "用户信息"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/users/me [put]
func (h *UserHandler) UpdateUserInfo(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "user"),
		zap.String("operation", "update_user_info"),
	)

	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		log.Warn("unauthorized",
			zap.String("event", "unauthorized"),
			zap.String("error_code", apperror.CodeUnauthorized),
			zap.String("error_kind", apperror.KindAuth),
		)
		response.Unauthorized(c, "invalid or expired token")
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		log.Warn("invalid_user_id",
			zap.String("event", "invalid_user_id"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeInvalidParams, apperror.KindValidation, "用户ID无效", http.StatusBadRequest))
		return
	}

	log = log.With(zap.Uint64("user_id", userID))

	var req UpdateUserInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("validation_failed",
			zap.String("event", "validation_failed"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeInvalidParams, apperror.KindValidation, "参数错误", http.StatusBadRequest))
		return
	}

	if err := h.service.UpdateUserInfo(ctx, userID, req.Nickname, req.Avatar); err != nil {
		log.Error("service_failed",
			zap.String("event", "service_failed"),
			zap.String("error_code", apperror.CodeUserUpdateFailed),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeUserUpdateFailed, apperror.KindDB, err.Error(), http.StatusInternalServerError))
		return
	}

	log.Info("user_info_updated",
		zap.String("event", "user_info_updated"),
		zap.String("entity_type", "user"),
		zap.Uint64("entity_id", userID),
	)
	platformlogger.Audit(c.Request.Context(), "user.update_info",
		zap.String("request_id", platformlogger.RequestIDFromGin(c)),
		zap.Uint64("user_id", userID),
		zap.String("entity_type", "user"),
		zap.Uint64("entity_id", userID),
		zap.String("result", "success"),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)

	response.SuccessWithMessage(c, "update success", nil)
}

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required" example:"123456"`
	NewPassword string `json:"newPassword" binding:"required,min=6,max=128" example:"654321"`
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改当前登录用户的密码
// @Tags 用户
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body ChangePasswordRequest true "密码信息"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/users/me/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "user"),
		zap.String("operation", "change_password"),
	)

	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		log.Warn("unauthorized",
			zap.String("event", "unauthorized"),
			zap.String("error_code", apperror.CodeUnauthorized),
			zap.String("error_kind", apperror.KindAuth),
		)
		response.Unauthorized(c, "invalid or expired token")
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		log.Warn("invalid_user_id",
			zap.String("event", "invalid_user_id"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeInvalidParams, apperror.KindValidation, "用户ID无效", http.StatusBadRequest))
		return
	}

	log = log.With(zap.Uint64("user_id", userID))

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("validation_failed",
			zap.String("event", "validation_failed"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeInvalidParams, apperror.KindValidation, "参数错误", http.StatusBadRequest))
		return
	}

	if err := h.service.ChangePassword(ctx, userID, req.OldPassword, req.NewPassword); err != nil {
		log.Error("service_failed",
			zap.String("event", "service_failed"),
			zap.String("error_code", apperror.CodeUserUpdateFailed),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeUserUpdateFailed, apperror.KindDB, err.Error(), http.StatusInternalServerError))
		return
	}

	log.Info("password_changed",
		zap.String("event", "password_changed"),
		zap.String("entity_type", "user"),
		zap.Uint64("entity_id", userID),
	)
	platformlogger.Audit(c.Request.Context(), "user.change_password",
		zap.String("request_id", platformlogger.RequestIDFromGin(c)),
		zap.Uint64("user_id", userID),
		zap.String("entity_type", "user"),
		zap.Uint64("entity_id", userID),
		zap.String("result", "success"),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)

	response.SuccessWithMessage(c, "password changed successfully", nil)
}

func ptrToStr(s *string) string { if s == nil { return "" }; return *s }
