package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/trailmemo/internal/middleware"
	"github.com/trailmemo/internal/model"
	platformlogger "github.com/trailmemo/internal/platform/logger"
	"github.com/trailmemo/internal/service"
	"github.com/trailmemo/pkg/apperror"
	"github.com/trailmemo/pkg/response"
)

type PostHandler struct {
	postService service.PostService
}

func NewPostHandler() *PostHandler {
	return &PostHandler{
		postService: service.NewPostService(),
	}
}

func (h *PostHandler) RegisterRoutes(r *gin.RouterGroup) {
	postGroup := r.Group("/posts")
	postGroup.Use(middleware.JWTAuth())
	{
		postGroup.POST("", h.CreatePost)
		postGroup.GET("", h.GetPostList)
		postGroup.GET("/:id", h.GetPostDetail)
		postGroup.PUT("/:id", h.UpdatePost)
		postGroup.DELETE("/:id", h.DeletePost)
	}
}

type CreatePostRequest struct {
	RouteID uint64 `json:"route_id" example:"1"`
	Title   string `json:"title" binding:"required" example:"我的旅行分享"`
	Content string `json:"content" binding:"required" example:"这次旅行太棒了！"`
	Images  string `json:"images" example:"https://example.com/photo1.jpg,https://example.com/photo2.jpg"`
}

// CreatePost 创建分享
// @Summary 创建分享
// @Description 创建新的旅行分享帖子
// @Tags 分享
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body CreatePostRequest true "分享信息"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/posts [post]
func (h *PostHandler) CreatePost(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "post"),
		zap.String("operation", "create_post"),
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

	var req CreatePostRequest
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

	post, err := h.postService.CreatePost(ctx, userID, req.RouteID, req.Title, req.Content, req.Images)
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

	log.Info("post_created",
		zap.String("event", "post_created"),
		zap.String("entity_type", "post"),
		zap.Uint64("entity_id", post.ID),
	)
	platformlogger.Audit(c.Request.Context(), "post.create",
		zap.String("request_id", platformlogger.RequestIDFromGin(c)),
		zap.Uint64("user_id", userID),
		zap.String("entity_type", "post"),
		zap.Uint64("entity_id", post.ID),
		zap.String("result", "success"),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)

	response.Success(c, post)
}

// GetPostList 获取分享列表
// @Summary 获取分享列表
// @Description 获取分享列表，可按用户或路线筛选
// @Tags 分享
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(20)
// @Param user_id query int false "用户ID"
// @Param route_id query int false "路线ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/posts [get]
func (h *PostHandler) GetPostList(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "post"),
		zap.String("operation", "list_posts"),
	)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	userIDStr := c.Query("user_id")
	routeIDStr := c.Query("route_id")

	log = log.With(
		zap.Int("page", page),
		zap.Int("size", size),
	)

	var posts []*model.Post
	var total int64
	var err error

	if userIDStr != "" {
		userID, parseErr := strconv.ParseUint(userIDStr, 10, 64)
		if parseErr != nil {
			log.Warn("validation_failed",
				zap.String("event", "validation_failed"),
				zap.String("error_code", apperror.CodeInvalidParams),
				zap.String("error_kind", apperror.KindValidation),
				zap.String("field", "user_id"),
				zap.Error(parseErr),
			)
			response.FromError(c, apperror.Wrap(parseErr, apperror.CodeInvalidParams, apperror.KindValidation, "用户ID参数错误", http.StatusBadRequest))
			return
		}
		log = log.With(zap.Uint64("user_id", userID))
		posts, total, err = h.postService.GetPostsByUserID(ctx, userID, page, size)
	} else if routeIDStr != "" {
		routeID, parseErr := strconv.ParseUint(routeIDStr, 10, 64)
		if parseErr != nil {
			log.Warn("validation_failed",
				zap.String("event", "validation_failed"),
				zap.String("error_code", apperror.CodeInvalidParams),
				zap.String("error_kind", apperror.KindValidation),
				zap.String("field", "route_id"),
				zap.Error(parseErr),
			)
			response.FromError(c, apperror.Wrap(parseErr, apperror.CodeInvalidParams, apperror.KindValidation, "路线ID参数错误", http.StatusBadRequest))
			return
		}
		log = log.With(zap.Uint64("route_id", routeID))
		posts, total, err = h.postService.GetPostsByRouteID(ctx, routeID, page, size)
	} else {
		posts, total, err = h.postService.GetAllPublicPosts(ctx, page, size)
	}

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
		"list":  posts,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// GetPostDetail 获取分享详情
// @Summary 获取分享详情
// @Description 获取指定ID的分享详情
// @Tags 分享
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "分享ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/posts/{id} [get]
func (h *PostHandler) GetPostDetail(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "post"),
		zap.String("operation", "get_post_detail"),
	)

	postIDStr := c.Param("id")
	postID, err := strconv.ParseUint(postIDStr, 10, 64)
	if err != nil {
		log.Warn("validation_failed",
			zap.String("event", "validation_failed"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.String("field", "id"),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeInvalidParams, apperror.KindValidation, "帖子ID参数错误", http.StatusBadRequest))
		return
	}

	log = log.With(zap.Uint64("post_id", postID))

	post, err := h.postService.GetPostByID(ctx, postID)
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
	if post == nil {
		log.Warn("post_not_found",
			zap.String("event", "post_not_found"),
		)
		response.NotFound(c, "post not found")
		return
	}

	response.Success(c, post)
}

type UpdatePostRequest struct {
	Title   string `json:"title" example:"更新后的标题"`
	Content string `json:"content" example:"更新后的内容"`
	Images  string `json:"images" example:"https://example.com/new-photo.jpg"`
}

// UpdatePost 更新分享
// @Summary 更新分享
// @Description 更新自己的分享内容
// @Tags 分享
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "分享ID"
// @Param body body UpdatePostRequest true "分享信息"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/posts/{id} [put]
func (h *PostHandler) UpdatePost(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "post"),
		zap.String("operation", "update_post"),
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

	postIDStr := c.Param("id")
	postID, err := strconv.ParseUint(postIDStr, 10, 64)
	if err != nil {
		log.Warn("validation_failed",
			zap.String("event", "validation_failed"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.String("field", "id"),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeInvalidParams, apperror.KindValidation, "帖子ID参数错误", http.StatusBadRequest))
		return
	}

	log = log.With(zap.Uint64("post_id", postID))

	var req UpdatePostRequest
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

	err = h.postService.UpdatePost(ctx, postID, userID, req.Title, req.Content, req.Images)
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

	log.Info("post_updated",
		zap.String("event", "post_updated"),
		zap.String("entity_type", "post"),
		zap.Uint64("entity_id", postID),
	)
	platformlogger.Audit(c.Request.Context(), "post.update",
		zap.String("request_id", platformlogger.RequestIDFromGin(c)),
		zap.Uint64("user_id", userID),
		zap.String("entity_type", "post"),
		zap.Uint64("entity_id", postID),
		zap.String("result", "success"),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)

	response.Success(c, "update success")
}

// DeletePost 删除分享
// @Summary 删除分享
// @Description 删除自己的分享
// @Tags 分享
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "分享ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/posts/{id} [delete]
func (h *PostHandler) DeletePost(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "post"),
		zap.String("operation", "delete_post"),
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

	postIDStr := c.Param("id")
	postID, err := strconv.ParseUint(postIDStr, 10, 64)
	if err != nil {
		log.Warn("validation_failed",
			zap.String("event", "validation_failed"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.String("field", "id"),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeInvalidParams, apperror.KindValidation, "帖子ID参数错误", http.StatusBadRequest))
		return
	}

	log = log.With(zap.Uint64("post_id", postID))

	err = h.postService.DeletePost(ctx, postID, userID)
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

	log.Info("post_deleted",
		zap.String("event", "post_deleted"),
		zap.String("entity_type", "post"),
		zap.Uint64("entity_id", postID),
	)
	platformlogger.Audit(c.Request.Context(), "post.delete",
		zap.String("request_id", platformlogger.RequestIDFromGin(c)),
		zap.Uint64("user_id", userID),
		zap.String("entity_type", "post"),
		zap.Uint64("entity_id", postID),
		zap.String("result", "success"),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)

	response.Success(c, "delete success")
}
