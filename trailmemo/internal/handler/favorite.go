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

type FavoriteHandler struct {
	favoriteService service.FavoriteService
}

func NewFavoriteHandler() *FavoriteHandler {
	return &FavoriteHandler{
		favoriteService: service.NewFavoriteService(),
	}
}

func (h *FavoriteHandler) RegisterRoutes(r *gin.RouterGroup) {
	favoriteGroup := r.Group("/favorites")
	favoriteGroup.Use(middleware.JWTAuth())
	{
		favoriteGroup.POST("/toggle", h.ToggleFavorite)
		favoriteGroup.GET("/status", h.GetFavoriteStatus)
		favoriteGroup.GET("/count", h.GetFavoriteCount)
		favoriteGroup.GET("/list", h.GetUserFavorites)
	}
}

type ToggleFavoriteRequest struct {
	RouteID uint64 `json:"route_id" binding:"required" example:"1"`
}

// ToggleFavorite 收藏/取消收藏
// @Summary 收藏/取消收藏
// @Description 对路线进行收藏或取消收藏
// @Tags 收藏
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body ToggleFavoriteRequest true "收藏信息"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/favorites/toggle [post]
func (h *FavoriteHandler) ToggleFavorite(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "favorite"),
		zap.String("operation", "toggle_favorite"),
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

	var req ToggleFavoriteRequest
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

	log = log.With(zap.Uint64("route_id", req.RouteID))

	favorited, err := h.favoriteService.ToggleFavorite(ctx, userID, req.RouteID)
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

	action := "unfavorite"
	if favorited {
		action = "favorite"
	}
	log.Info("favorite_toggled",
		zap.String("event", "favorite_toggled"),
		zap.String("entity_type", "favorite"),
		zap.Uint64("user_id", userID),
		zap.Uint64("route_id", req.RouteID),
		zap.Bool("is_favorited", favorited),
	)
	platformlogger.Audit(c.Request.Context(), "favorite.toggle",
		zap.String("request_id", platformlogger.RequestIDFromGin(c)),
		zap.Uint64("user_id", userID),
		zap.String("entity_type", "favorite"),
		zap.Uint64("route_id", req.RouteID),
		zap.String("action", action),
		zap.Bool("result", favorited),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)

	response.Success(c, gin.H{
		"favorited": favorited,
	})
}

// GetFavoriteStatus 获取收藏状态
// @Summary 获取收藏状态
// @Description 检查用户是否已收藏指定路线
// @Tags 收藏
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param route_id query int true "路线ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/favorites/status [get]
func (h *FavoriteHandler) GetFavoriteStatus(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "favorite"),
		zap.String("operation", "get_favorite_status"),
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

	routeIDStr := c.Query("route_id")
	if routeIDStr == "" {
		log.Warn("validation_failed",
			zap.String("event", "validation_failed"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
		)
		response.FromError(c, apperror.New(apperror.CodeInvalidParams, apperror.KindValidation, "route_id is required", http.StatusBadRequest))
		return
	}
	routeID, err := strconv.ParseUint(routeIDStr, 10, 64)
	if err != nil {
		log.Warn("invalid_route_id",
			zap.String("event", "invalid_route_id"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeInvalidParams, apperror.KindValidation, "路线ID无效", http.StatusBadRequest))
		return
	}

	log = log.With(zap.Uint64("user_id", userID), zap.Uint64("route_id", routeID))

	favorited, err := h.favoriteService.CheckFavoriteStatus(ctx, userID, routeID)
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

	log.Info("favorite_status_checked",
		zap.String("event", "favorite_status_checked"),
		zap.Bool("is_favorited", favorited),
	)

	response.Success(c, gin.H{
		"favorited": favorited,
	})
}

// GetFavoriteCount 获取收藏数量
// @Summary 获取收藏数量
// @Description 获取路线的收藏数量
// @Tags 收藏
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param route_id query int true "路线ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/favorites/count [get]
func (h *FavoriteHandler) GetFavoriteCount(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "favorite"),
		zap.String("operation", "get_favorite_count"),
	)

	routeIDStr := c.Query("route_id")
	if routeIDStr == "" {
		log.Warn("validation_failed",
			zap.String("event", "validation_failed"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
		)
		response.FromError(c, apperror.New(apperror.CodeInvalidParams, apperror.KindValidation, "route_id is required", http.StatusBadRequest))
		return
	}
	routeID, err := strconv.ParseUint(routeIDStr, 10, 64)
	if err != nil {
		log.Warn("invalid_route_id",
			zap.String("event", "invalid_route_id"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeInvalidParams, apperror.KindValidation, "路线ID无效", http.StatusBadRequest))
		return
	}

	log = log.With(zap.Uint64("route_id", routeID))

	count, err := h.favoriteService.GetFavoriteCount(ctx, routeID)
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

	log.Info("favorite_count_retrieved",
		zap.String("event", "favorite_count_retrieved"),
		zap.Int64("count", count),
	)

	response.Success(c, gin.H{
		"count": count,
	})
}

// GetUserFavorites 获取用户收藏列表
// @Summary 获取用户收藏列表
// @Description 获取用户收藏的所有路线
// @Tags 收藏
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(20)
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/favorites/list [get]
func (h *FavoriteHandler) GetUserFavorites(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "favorite"),
		zap.String("operation", "get_user_favorites"),
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

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	favorites, total, err := h.favoriteService.GetUserFavorites(ctx, userID, page, size)
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

	log.Info("user_favorites_retrieved",
		zap.String("event", "user_favorites_retrieved"),
		zap.Int64("total", total),
	)

	response.Success(c, gin.H{
		"list":  favorites,
		"total": total,
		"page":  page,
		"size":  size,
	})
}
