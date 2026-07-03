package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trailmemo/internal/middleware"
	platformlogger "github.com/trailmemo/internal/platform/logger"
	"github.com/trailmemo/internal/service"
	"github.com/trailmemo/pkg/apperror"
	"github.com/trailmemo/pkg/response"
	"go.uber.org/zap"
)

type LikeHandler struct {
	likeService service.LikeService
}

func NewLikeHandler() *LikeHandler {
	return &LikeHandler{
		likeService: service.NewLikeService(),
	}
}

func (h *LikeHandler) RegisterRoutes(r *gin.RouterGroup) {
	likeGroup := r.Group("/likes")
	likeGroup.Use(middleware.JWTAuth())
	{
		likeGroup.POST("/toggle", h.ToggleLike)   // 点赞/取消点赞
		likeGroup.GET("/status", h.GetLikeStatus) // 获取点赞状态
		likeGroup.GET("/count", h.GetLikeCount)   // 获取点赞数量
	}
}

type ToggleLikeRequest struct { // 点赞/取消点赞请求
	TargetID   uint64 `json:"target_id" binding:"required" example:"1"`
	TargetType string `json:"target_type" binding:"required" example:"post"`
}

// ToggleLike 点赞/取消点赞
// @Summary 点赞/取消点赞
// @Description 对分享或评论进行点赞或取消点赞
// @Tags 点赞
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body ToggleLikeRequest true "点赞信息"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/likes/toggle [post]
func (h *LikeHandler) ToggleLike(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "like"),
		zap.String("operation", "toggle_like"),
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

	var req ToggleLikeRequest
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

	if req.TargetType != "post" && req.TargetType != "comment" {
		log.Warn("invalid_target_type",
			zap.String("event", "invalid_target_type"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.String("target_type", req.TargetType),
		)
		response.FromError(c, apperror.New(apperror.CodeInvalidParams, apperror.KindValidation, "invalid target_type, must be 'post' or 'comment'", http.StatusBadRequest))
		return
	}

	log = log.With(zap.Uint64("target_id", req.TargetID), zap.String("target_type", req.TargetType))

	liked, err := h.likeService.ToggleLike(ctx, userID, req.TargetID, req.TargetType)
	if err != nil {
		appErr := apperror.From(err)
		log.Error("service_failed",
			zap.String("event", "service_failed"),
			zap.String("error_code", appErr.Code),
			zap.String("error_kind", appErr.Kind),
			zap.Error(err),
		)
		response.FromError(c, err)
		return
	}

	action := "unlike"
	if liked {
		action = "like"
	}
	log.Info("like_toggled",
		zap.String("event", "like_toggled"),
		zap.String("entity_type", "like"),
		zap.Uint64("user_id", userID),
		zap.Uint64("target_id", req.TargetID),
		zap.String("target_type", req.TargetType),
		zap.Bool("is_liked", liked),
	)
	platformlogger.Audit(c.Request.Context(), "like.toggle",
		zap.String("request_id", platformlogger.RequestIDFromGin(c)),
		zap.Uint64("user_id", userID),
		zap.String("entity_type", "like"),
		zap.Uint64("target_id", req.TargetID),
		zap.String("target_type", req.TargetType),
		zap.String("action", action),
		zap.Bool("result", liked),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)

	response.Success(c, gin.H{
		"liked": liked,
	})
}

// GetLikeStatus 获取点赞状态
// @Summary 获取点赞状态
// @Description 检查用户是否已对目标进行点赞
// @Tags 点赞
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param target_id query int true "目标ID"
// @Param target_type query string true "目标类型"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/likes/status [get]
func (h *LikeHandler) GetLikeStatus(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "like"),
		zap.String("operation", "get_like_status"),
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

	targetIDStr := c.Query("target_id")
	if targetIDStr == "" {
		log.Warn("validation_failed",
			zap.String("event", "validation_failed"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
		)
		response.FromError(c, apperror.New(apperror.CodeInvalidParams, apperror.KindValidation, "target_id is required", http.StatusBadRequest))
		return
	}
	targetID, err := strconv.ParseUint(targetIDStr, 10, 64)
	if err != nil {
		log.Warn("invalid_target_id",
			zap.String("event", "invalid_target_id"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeInvalidParams, apperror.KindValidation, "目标ID无效", http.StatusBadRequest))
		return
	}

	targetType := c.Query("target_type")
	if targetType == "" {
		log.Warn("validation_failed",
			zap.String("event", "validation_failed"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
		)
		response.FromError(c, apperror.New(apperror.CodeInvalidParams, apperror.KindValidation, "target_type is required", http.StatusBadRequest))
		return
	}
	if targetType != "post" && targetType != "comment" {
		log.Warn("invalid_target_type",
			zap.String("event", "invalid_target_type"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.String("target_type", targetType),
		)
		response.FromError(c, apperror.New(apperror.CodeInvalidParams, apperror.KindValidation, "invalid target_type, must be 'post' or 'comment'", http.StatusBadRequest))
		return
	}

	log = log.With(zap.Uint64("user_id", userID), zap.Uint64("target_id", targetID), zap.String("target_type", targetType))

	liked, err := h.likeService.CheckLikeStatus(ctx, userID, targetID, targetType)
	if err != nil {
		appErr := apperror.From(err)
		log.Error("service_failed",
			zap.String("event", "service_failed"),
			zap.String("error_code", appErr.Code),
			zap.String("error_kind", appErr.Kind),
			zap.Error(err),
		)
		response.FromError(c, err)
		return
	}

	log.Info("like_status_checked",
		zap.String("event", "like_status_checked"),
		zap.Bool("is_liked", liked),
	)

	response.Success(c, gin.H{
		"liked": liked,
	})
}

// GetLikeCount 获取点赞数量
// @Summary 获取点赞数量
// @Description 获取目标内容的点赞数量
// @Tags 点赞
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param target_id query int true "目标ID"
// @Param target_type query string true "目标类型"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/likes/count [get]
func (h *LikeHandler) GetLikeCount(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "like"),
		zap.String("operation", "get_like_count"),
	)

	targetIDStr := c.Query("target_id")
	if targetIDStr == "" {
		log.Warn("validation_failed",
			zap.String("event", "validation_failed"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
		)
		response.FromError(c, apperror.New(apperror.CodeInvalidParams, apperror.KindValidation, "target_id is required", http.StatusBadRequest))
		return
	}
	targetID, err := strconv.ParseUint(targetIDStr, 10, 64)
	if err != nil {
		log.Warn("invalid_target_id",
			zap.String("event", "invalid_target_id"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeInvalidParams, apperror.KindValidation, "目标ID无效", http.StatusBadRequest))
		return
	}

	targetType := c.Query("target_type")
	if targetType == "" {
		log.Warn("validation_failed",
			zap.String("event", "validation_failed"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
		)
		response.FromError(c, apperror.New(apperror.CodeInvalidParams, apperror.KindValidation, "target_type is required", http.StatusBadRequest))
		return
	}
	if targetType != "post" && targetType != "comment" {
		log.Warn("invalid_target_type",
			zap.String("event", "invalid_target_type"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.String("target_type", targetType),
		)
		response.FromError(c, apperror.New(apperror.CodeInvalidParams, apperror.KindValidation, "invalid target_type, must be 'post' or 'comment'", http.StatusBadRequest))
		return
	}

	log = log.With(zap.Uint64("target_id", targetID), zap.String("target_type", targetType))

	count, err := h.likeService.GetLikeCount(ctx, targetID, targetType)
	if err != nil {
		appErr := apperror.From(err)
		log.Error("service_failed",
			zap.String("event", "service_failed"),
			zap.String("error_code", appErr.Code),
			zap.String("error_kind", appErr.Kind),
			zap.Error(err),
		)
		response.FromError(c, err)
		return
	}

	log.Info("like_count_retrieved",
		zap.String("event", "like_count_retrieved"),
		zap.Int64("count", count),
	)

	response.Success(c, gin.H{
		"count": count,
	})
}
