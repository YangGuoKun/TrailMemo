package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trailmemo/internal/middleware"
	"github.com/trailmemo/internal/model"
	platformlogger "github.com/trailmemo/internal/platform/logger"
	"github.com/trailmemo/internal/service"
	"github.com/trailmemo/pkg/apperror"
	"github.com/trailmemo/pkg/response"
	"go.uber.org/zap"
)

type CheckinHandler struct {
	checkinService service.CheckinService
}

func NewCheckinHandler() *CheckinHandler {
	return &CheckinHandler{
		checkinService: service.NewCheckinService(),
	}
}

func (h *CheckinHandler) RegisterRoutes(r *gin.RouterGroup) {
	checkinGroup := r.Group("/checkins")
	checkinGroup.Use(middleware.JWTAuth())
	{
		checkinGroup.POST("", h.CreateCheckin)
		checkinGroup.GET("", h.GetCheckinList)
		checkinGroup.GET("/:id", h.GetCheckinDetail)
		checkinGroup.PUT("/:id", h.UpdateCheckin)
		checkinGroup.DELETE("/:id", h.DeleteCheckin)
		checkinGroup.GET("/progress/:route_id", h.GetRouteProgress)
	}
}

type CreateCheckinRequest struct {
	RouteID      uint64  `json:"route_id" binding:"required" example:"1"`
	CheckpointID uint64  `json:"checkpoint_id" binding:"required" example:"1"`
	Latitude     float64 `json:"latitude" example:"30.6562"`
	Longitude    float64 `json:"longitude" example:"104.0657"`
	PhotoURL     string  `json:"photo_url" example:"https://example.com/checkin.jpg"`
	Content      string  `json:"content" example:"这里的美食太好吃了！"`
	Rating       int     `json:"rating" binding:"min=1,max=5" example:"5"`
}

// CreateCheckin 创建打卡记录
// @Summary 创建打卡记录
// @Description 在指定打卡点创建新的打卡记录
// @Tags 打卡
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body CreateCheckinRequest true "打卡信息"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/checkins [post]
func (h *CheckinHandler) CreateCheckin(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "checkin"),
		zap.String("operation", "create_checkin"),
	)

	userID, ok := h.currentUserID(c, log)
	if !ok {
		return
	}
	log = log.With(zap.Uint64("user_id", userID))

	var req CreateCheckinRequest
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
	log = log.With(
		zap.Uint64("route_id", req.RouteID),
		zap.Uint64("checkpoint_id", req.CheckpointID),
	)

	checkin, err := h.checkinService.CreateCheckin(
		ctx,
		userID,
		req.RouteID,
		req.CheckpointID,
		req.Latitude,
		req.Longitude,
		req.PhotoURL,
		req.Content,
		req.Rating,
	)
	if err != nil {
		h.writeServiceError(c, log, err)
		return
	}

	log.Info("checkin_created",
		zap.String("event", "checkin_created"),
		zap.String("entity_type", "checkin"),
		zap.Uint64("entity_id", checkin.ID),
		zap.Uint64("route_id", checkin.RouteID),
		zap.Uint64("checkpoint_id", checkin.CheckpointID),
	)
	platformlogger.Audit(c.Request.Context(), "checkin.create",
		zap.String("request_id", platformlogger.RequestIDFromGin(c)),
		zap.Uint64("user_id", userID),
		zap.String("entity_type", "checkin"),
		zap.Uint64("entity_id", checkin.ID),
		zap.Uint64("route_id", checkin.RouteID),
		zap.Uint64("checkpoint_id", checkin.CheckpointID),
		zap.String("result", "success"),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)
	response.Success(c, checkin)
}

// GetCheckinList 获取打卡记录列表
// @Summary 获取打卡记录列表
// @Description 获取用户的所有打卡记录，可按路线筛选
// @Tags 打卡
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(20)
// @Param route_id query int false "路线ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/checkins [get]
func (h *CheckinHandler) GetCheckinList(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "checkin"),
		zap.String("operation", "list_checkins"),
	)

	userID, ok := h.currentUserID(c, log)
	if !ok {
		return
	}
	log = log.With(zap.Uint64("user_id", userID))

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	routeIDStr := c.Query("route_id")

	var checkins []*model.Checkin
	var total int64
	var err error

	if routeIDStr != "" {
		routeID, parseErr := strconv.ParseUint(routeIDStr, 10, 64)
		if parseErr != nil {
			log.Warn("validation_failed",
				zap.String("event", "validation_failed"),
				zap.String("error_code", apperror.CodeInvalidParams),
				zap.String("error_kind", apperror.KindValidation),
				zap.String("field", "route_id"),
				zap.Error(parseErr),
			)
			response.FromError(c, apperror.Wrap(parseErr, apperror.CodeInvalidParams, apperror.KindValidation, "路线 ID 参数错误", http.StatusBadRequest))
			return
		}
		log = log.With(zap.Uint64("route_id", routeID))
		checkins, total, err = h.checkinService.GetCheckinsByRouteID(ctx, routeID, page, size)
	} else {
		checkins, total, err = h.checkinService.GetCheckinsByUserID(ctx, userID, page, size)
	}

	if err != nil {
		h.writeServiceError(c, log, err)
		return
	}

	response.Success(c, gin.H{
		"list":  checkins,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// GetCheckinDetail 获取打卡详情
// @Summary 获取打卡详情
// @Description 获取指定ID的打卡记录详情
// @Tags 打卡
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "打卡记录ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/checkins/{id} [get]
func (h *CheckinHandler) GetCheckinDetail(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "checkin"),
		zap.String("operation", "get_checkin_detail"),
	)

	checkinID, ok := h.parseUintParam(c, log, "id", "打卡记录 ID 参数错误")
	if !ok {
		return
	}
	log = log.With(zap.Uint64("checkin_id", checkinID))

	checkin, err := h.checkinService.GetCheckinByID(ctx, checkinID)
	if err != nil {
		h.writeServiceError(c, log, err)
		return
	}
	if checkin == nil {
		response.FromError(c, apperror.New(apperror.CodeCheckinNotFound, apperror.KindNotFound, "打卡记录不存在", http.StatusNotFound))
		return
	}

	response.Success(c, checkin)
}

type UpdateCheckinRequest struct {
	PhotoURL string `json:"photo_url" example:"https://example.com/new-photo.jpg"`
	Content  string `json:"content" example:"更新后的内容"`
	Rating   int    `json:"rating" binding:"min=1,max=5" example:"4"`
}

// UpdateCheckin 更新打卡记录
// @Summary 更新打卡记录
// @Description 更新自己的打卡记录
// @Tags 打卡
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "打卡记录ID"
// @Param body body UpdateCheckinRequest true "打卡信息"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/checkins/{id} [put]
func (h *CheckinHandler) UpdateCheckin(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "checkin"),
		zap.String("operation", "update_checkin"),
	)

	userID, ok := h.currentUserID(c, log)
	if !ok {
		return
	}
	log = log.With(zap.Uint64("user_id", userID))

	checkinID, ok := h.parseUintParam(c, log, "id", "打卡记录 ID 参数错误")
	if !ok {
		return
	}
	log = log.With(zap.Uint64("checkin_id", checkinID))

	var req UpdateCheckinRequest
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

	if err := h.checkinService.UpdateCheckin(ctx, checkinID, userID, req.PhotoURL, req.Content, req.Rating); err != nil {
		h.writeServiceError(c, log, err)
		return
	}

	log.Info("checkin_updated",
		zap.String("event", "checkin_updated"),
		zap.String("entity_type", "checkin"),
		zap.Uint64("entity_id", checkinID),
	)
	platformlogger.Audit(c.Request.Context(), "checkin.update",
		zap.String("request_id", platformlogger.RequestIDFromGin(c)),
		zap.Uint64("user_id", userID),
		zap.String("entity_type", "checkin"),
		zap.Uint64("entity_id", checkinID),
		zap.String("result", "success"),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)
	response.Success(c, "update success")
}

// DeleteCheckin 删除打卡记录
// @Summary 删除打卡记录
// @Description 删除自己的打卡记录
// @Tags 打卡
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "打卡记录ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/checkins/{id} [delete]
func (h *CheckinHandler) DeleteCheckin(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "checkin"),
		zap.String("operation", "delete_checkin"),
	)

	userID, ok := h.currentUserID(c, log)
	if !ok {
		return
	}
	log = log.With(zap.Uint64("user_id", userID))

	checkinID, ok := h.parseUintParam(c, log, "id", "打卡记录 ID 参数错误")
	if !ok {
		return
	}
	log = log.With(zap.Uint64("checkin_id", checkinID))

	if err := h.checkinService.DeleteCheckin(ctx, checkinID, userID); err != nil {
		h.writeServiceError(c, log, err)
		return
	}

	log.Info("checkin_deleted",
		zap.String("event", "checkin_deleted"),
		zap.String("entity_type", "checkin"),
		zap.Uint64("entity_id", checkinID),
	)
	platformlogger.Audit(c.Request.Context(), "checkin.delete",
		zap.String("request_id", platformlogger.RequestIDFromGin(c)),
		zap.Uint64("user_id", userID),
		zap.String("entity_type", "checkin"),
		zap.Uint64("entity_id", checkinID),
		zap.String("result", "success"),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)
	response.Success(c, "delete success")
}

// GetRouteProgress 获取路线打卡进度
// @Summary 获取路线打卡进度
// @Description 获取指定路线的打卡进度统计
// @Tags 打卡
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param route_id path int true "路线ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/checkins/progress/{route_id} [get]
func (h *CheckinHandler) GetRouteProgress(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "checkin"),
		zap.String("operation", "get_route_progress"),
	)

	userID, ok := h.currentUserID(c, log)
	if !ok {
		return
	}
	log = log.With(zap.Uint64("user_id", userID))

	routeID, ok := h.parseUintParam(c, log, "route_id", "路线 ID 参数错误")
	if !ok {
		return
	}
	log = log.With(zap.Uint64("route_id", routeID))

	progress, err := h.checkinService.GetRouteProgress(ctx, userID, routeID)
	if err != nil {
		h.writeServiceError(c, log, err)
		return
	}

	response.Success(c, gin.H{
		"progress": progress,
	})
}

func (h *CheckinHandler) currentUserID(c *gin.Context, log *zap.Logger) (uint64, bool) {
	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		response.Unauthorized(c, "invalid or expired token")
		return 0, false
	}
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		log.Warn("validation_failed",
			zap.String("event", "validation_failed"),
			zap.String("error_code", apperror.CodeUnauthorized),
			zap.String("error_kind", apperror.KindAuth),
			zap.String("field", "user_id"),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeUnauthorized, apperror.KindAuth, "用户身份无效", http.StatusUnauthorized))
		return 0, false
	}
	return userID, true
}

func (h *CheckinHandler) parseUintParam(c *gin.Context, log *zap.Logger, name, message string) (uint64, bool) {
	value, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		log.Warn("validation_failed",
			zap.String("event", "validation_failed"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.String("field", name),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeInvalidParams, apperror.KindValidation, message, http.StatusBadRequest))
		return 0, false
	}
	return value, true
}

func (h *CheckinHandler) writeServiceError(c *gin.Context, log *zap.Logger, err error) {
	appErr := apperror.From(err)
	log.Error("service_failed",
		zap.String("event", "service_failed"),
		zap.String("error_code", appErr.Code),
		zap.String("error_kind", appErr.Kind),
		zap.Error(err),
	)
	response.FromError(c, err)
}
