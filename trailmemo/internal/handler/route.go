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

type RouteHandler struct {
	routeService service.RouteService
}

func NewRouteHandler() *RouteHandler {
	return &RouteHandler{
		routeService: service.NewRouteService(),
	}
}

func (h *RouteHandler) RegisterRoutes(r *gin.RouterGroup) {
	routeGroup := r.Group("/routes")
	routeGroup.Use(middleware.JWTAuth())
	{
		routeGroup.POST("", h.CreateRoute)
		routeGroup.GET("", h.GetRouteList)
		routeGroup.GET("/:id", h.GetRouteDetail)
		routeGroup.PUT("/:id", h.UpdateRoute)
		routeGroup.DELETE("/:id", h.DeleteRoute)
		routeGroup.POST("/:id/copy", h.CopyRoute)
	}
}

type CreateRouteRequest struct {
	Title          string             `json:"title" binding:"required" example:"成都三日游"`
	Description    string             `json:"description" example:"成都美食之旅"`
	CoverImage     string             `json:"coverImage" example:"https://example.com/cover.jpg"`
	StartCity      string             `json:"startCity" binding:"required" example:"成都"`
	EndCity        string             `json:"endCity" binding:"required" example:"成都"`
	TotalDistance  float64            `json:"totalDistance" example:"100.5"`
	EstimatedHours float64            `json:"estimatedHours" example:"24.0"`
	IsPublic       int                `json:"isPublic" example:"1"`
	Checkpoints    []*CheckpointInput `json:"checkpoints"`
}

type CheckpointInput struct {
	Name         string  `json:"name" binding:"required" example:"宽窄巷子"`
	Description  string  `json:"description" example:"成都著名景点"`
	Latitude     float64 `json:"latitude" example:"30.6562"`
	Longitude    float64 `json:"longitude" example:"104.0657"`
	Address      string  `json:"address" example:"成都市青羊区宽窄巷子"`
	City         string  `json:"city" example:"成都"`
	Sequence     int     `json:"sequence" binding:"required" example:"1"`
	ArriveTime   string  `json:"arriveTime" example:"2024-06-01 10:00"`
	StayDuration int     `json:"stayDuration" example:"120"`
	PhotoURL     string  `json:"photoURL" example:"https://example.com/photo.jpg"`
}

// CreateRoute 创建路线
// @Summary 创建路线
// @Description 创建新的旅游路线，包含多个打卡点
// @Tags 路线
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body CreateRouteRequest true "路线信息"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/routes [post]
func (h *RouteHandler) CreateRoute(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "route"),
		zap.String("operation", "create_route"),
	)

	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		response.Unauthorized(c, "invalid or expired token")
		return
	}
	userID, _ := strconv.ParseUint(userIDStr, 10, 64)
	log = log.With(zap.Uint64("user_id", userID))

	var req CreateRouteRequest
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

	if req.IsPublic == 0 {
		req.IsPublic = 1
	}

	checkpoints := make([]*model.Checkpoint, 0)
	for _, cp := range req.Checkpoints {
		checkpoints = append(checkpoints, &model.Checkpoint{
			Name:         cp.Name,
			Description:  cp.Description,
			Latitude:     cp.Latitude,
			Longitude:    cp.Longitude,
			Address:      cp.Address,
			City:         cp.City,
			Sequence:     cp.Sequence,
			ArriveTime:   cp.ArriveTime,
			StayDuration: cp.StayDuration,
			PhotoURL:     cp.PhotoURL,
		})
	}

	route, err := h.routeService.CreateRoute(
		ctx,
		userID,
		req.Title,
		req.Description,
		req.CoverImage,
		req.StartCity,
		req.EndCity,
		req.TotalDistance,
		req.EstimatedHours,
		req.IsPublic,
		checkpoints,
	)
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

	log.Info("route_created",
		zap.String("event", "route_created"),
		zap.String("entity_type", "route"),
		zap.Uint64("entity_id", route.ID),
		zap.Int("checkpoint_count", len(req.Checkpoints)),
		zap.String("start_city", req.StartCity),
		zap.String("end_city", req.EndCity),
	)
	platformlogger.Audit(c.Request.Context(), "route.create",
		zap.String("request_id", platformlogger.RequestIDFromGin(c)),
		zap.Uint64("user_id", userID),
		zap.String("entity_type", "route"),
		zap.Uint64("entity_id", route.ID),
		zap.String("result", "success"),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)
	response.Success(c, route)
}

// GetRouteList 获取路线列表
// @Summary 获取路线列表
// @Description 获取当前用户创建的所有路线
// @Tags 路线
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(20)
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/routes [get]
func (h *RouteHandler) GetRouteList(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "route"),
		zap.String("operation", "list_routes"),
	)

	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		response.Unauthorized(c, "invalid or expired token")
		return
	}
	userID, _ := strconv.ParseUint(userIDStr, 10, 64)
	log = log.With(zap.Uint64("user_id", userID))

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	routes, total, err := h.routeService.GetRoutesByUserID(ctx, userID, page, size)
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

	response.Success(c, gin.H{
		"list":  routes,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// GetRouteDetail 获取路线详情
// @Summary 获取路线详情
// @Description 获取指定ID的路线详细信息，包含所有打卡点
// @Tags 路线
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "路线ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/routes/{id} [get]
func (h *RouteHandler) GetRouteDetail(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "route"),
		zap.String("operation", "get_route_detail"),
	)

	routeIDStr := c.Param("id")
	routeID, err := strconv.ParseUint(routeIDStr, 10, 64)
	if err != nil {
		log.Warn("validation_failed",
			zap.String("event", "validation_failed"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.String("field", "id"),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeInvalidParams, apperror.KindValidation, "路线 ID 参数错误", http.StatusBadRequest))
		return
	}
	log = log.With(zap.Uint64("route_id", routeID))

	route, err := h.routeService.GetRouteByID(ctx, routeID)
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
	if route == nil {
		response.FromError(c, apperror.New(apperror.CodeRouteNotFound, apperror.KindNotFound, "路线不存在", http.StatusNotFound))
		return
	}

	response.Success(c, route)
}

type UpdateRouteRequest struct {
	Title          string  `json:"title" example:"成都三日游（更新）"`
	Description    string  `json:"description" example:"更新后的描述"`
	CoverImage     string  `json:"coverImage" example:"https://example.com/new-cover.jpg"`
	StartCity      string  `json:"startCity" example:"成都"`
	EndCity        string  `json:"endCity" example:"成都"`
	TotalDistance  float64 `json:"totalDistance" example:"150.0"`
	EstimatedHours float64 `json:"estimatedHours" example:"36.0"`
	IsPublic       int     `json:"isPublic" example:"1"`
}

// UpdateRoute 更新路线
// @Summary 更新路线
// @Description 更新自己创建的路线信息
// @Tags 路线
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "路线ID"
// @Param body body UpdateRouteRequest true "路线信息"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/routes/{id} [put]
func (h *RouteHandler) UpdateRoute(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "route"),
		zap.String("operation", "update_route"),
	)

	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		response.Unauthorized(c, "invalid or expired token")
		return
	}
	userID, _ := strconv.ParseUint(userIDStr, 10, 64)
	log = log.With(zap.Uint64("user_id", userID))

	routeIDStr := c.Param("id")
	routeID, err := strconv.ParseUint(routeIDStr, 10, 64)
	if err != nil {
		log.Warn("validation_failed",
			zap.String("event", "validation_failed"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.String("field", "id"),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeInvalidParams, apperror.KindValidation, "路线 ID 参数错误", http.StatusBadRequest))
		return
	}
	log = log.With(zap.Uint64("route_id", routeID))

	var req UpdateRouteRequest
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

	err = h.routeService.UpdateRoute(
		ctx,
		routeID,
		userID,
		req.Title,
		req.Description,
		req.CoverImage,
		req.StartCity,
		req.EndCity,
		req.TotalDistance,
		req.EstimatedHours,
		req.IsPublic,
	)
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

	log.Info("route_updated",
		zap.String("event", "route_updated"),
		zap.String("entity_type", "route"),
		zap.Uint64("entity_id", routeID),
	)
	platformlogger.Audit(c.Request.Context(), "route.update",
		zap.String("request_id", platformlogger.RequestIDFromGin(c)),
		zap.Uint64("user_id", userID),
		zap.String("entity_type", "route"),
		zap.Uint64("entity_id", routeID),
		zap.String("result", "success"),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)
	response.Success(c, "update success")
}

// DeleteRoute 删除路线
// @Summary 删除路线
// @Description 删除自己创建的路线
// @Tags 路线
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "路线ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/routes/{id} [delete]
func (h *RouteHandler) DeleteRoute(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "route"),
		zap.String("operation", "delete_route"),
	)

	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		response.Unauthorized(c, "invalid or expired token")
		return
	}
	userID, _ := strconv.ParseUint(userIDStr, 10, 64)
	log = log.With(zap.Uint64("user_id", userID))

	routeIDStr := c.Param("id")
	routeID, err := strconv.ParseUint(routeIDStr, 10, 64)
	if err != nil {
		log.Warn("validation_failed",
			zap.String("event", "validation_failed"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.String("field", "id"),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeInvalidParams, apperror.KindValidation, "路线 ID 参数错误", http.StatusBadRequest))
		return
	}
	log = log.With(zap.Uint64("route_id", routeID))

	err = h.routeService.DeleteRoute(ctx, routeID, userID)
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

	log.Info("route_deleted",
		zap.String("event", "route_deleted"),
		zap.String("entity_type", "route"),
		zap.Uint64("entity_id", routeID),
	)
	platformlogger.Audit(c.Request.Context(), "route.delete",
		zap.String("request_id", platformlogger.RequestIDFromGin(c)),
		zap.Uint64("user_id", userID),
		zap.String("entity_type", "route"),
		zap.Uint64("entity_id", routeID),
		zap.String("result", "success"),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)
	response.Success(c, "delete success")
}

type CopyRouteRequest struct {
	IsPublic int `json:"isPublic" example:"1"`
}

// CopyRoute 复用路线
// @Summary 复用路线
// @Description 复制他人公开的路线，成为自己的路线
// @Tags 路线
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "原路线ID"
// @Param body body CopyRouteRequest false "新路线公开状态"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/routes/{id}/copy [post]
func (h *RouteHandler) CopyRoute(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "route"),
		zap.String("operation", "copy_route"),
	)

	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		response.Unauthorized(c, "invalid or expired token")
		return
	}
	userID, _ := strconv.ParseUint(userIDStr, 10, 64)
	log = log.With(zap.Uint64("user_id", userID))

	routeIDStr := c.Param("id")
	routeID, err := strconv.ParseUint(routeIDStr, 10, 64)
	if err != nil {
		log.Warn("validation_failed",
			zap.String("event", "validation_failed"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.String("field", "id"),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeInvalidParams, apperror.KindValidation, "路线 ID 参数错误", http.StatusBadRequest))
		return
	}
	log = log.With(zap.Uint64("route_id", routeID))

	var req CopyRouteRequest
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

	isPublic := req.IsPublic
	if isPublic != 1 {
		isPublic = 1 // 默认公开
	}

	newRoute, err := h.routeService.CopyRoute(ctx, userID, routeID, isPublic)
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

	log.Info("route_copied",
		zap.String("event", "route_copied"),
		zap.String("entity_type", "route"),
		zap.Uint64("entity_id", newRoute.ID),
		zap.Uint64("source_route_id", routeID),
	)
	platformlogger.Audit(c.Request.Context(), "route.copy",
		zap.String("request_id", platformlogger.RequestIDFromGin(c)),
		zap.Uint64("user_id", userID),
		zap.String("entity_type", "route"),
		zap.Uint64("entity_id", newRoute.ID),
		zap.Uint64("source_route_id", routeID),
		zap.String("result", "success"),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)
	response.Success(c, newRoute)
}
