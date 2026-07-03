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

type CommentHandler struct {
	commentService service.CommentService
	likeService    service.LikeService
}

func NewCommentHandler() *CommentHandler {
	return &CommentHandler{
		commentService: service.NewCommentService(),
		likeService:    service.NewLikeService(),
	}
}

func (h *CommentHandler) RegisterRoutes(r *gin.RouterGroup) {
	commentGroup := r.Group("/comments")
	commentGroup.Use(middleware.JWTAuth())
	{
		commentGroup.POST("", h.CreateComment)
		commentGroup.GET("", h.GetCommentList)
		commentGroup.GET("/:id", h.GetCommentDetail)
		commentGroup.PUT("/:id", h.UpdateComment)
		commentGroup.DELETE("/:id", h.DeleteComment)
	}
}

type CreateCommentRequest struct {
	PostID   uint64 `json:"post_id" binding:"required" example:"1"`
	ParentID uint64 `json:"parent_id" example:"0"`
	Content  string `json:"content" binding:"required" example:"评论内容"`
}

// CreateComment 创建评论
// @Summary 创建评论
// @Description 在分享下创建评论
// @Tags 评论
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body CreateCommentRequest true "评论信息"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/comments [post]
func (h *CommentHandler) CreateComment(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "comment"),
		zap.String("operation", "create_comment"),
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

	var req CreateCommentRequest
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

	log = log.With(zap.Uint64("post_id", req.PostID))

	comment, err := h.commentService.CreateComment(ctx, userID, req.PostID, req.ParentID, req.Content)
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

	log.Info("comment_created",
		zap.String("event", "comment_created"),
		zap.String("entity_type", "comment"),
		zap.Uint64("entity_id", comment.ID),
	)
	platformlogger.Audit(c.Request.Context(), "comment.create",
		zap.String("request_id", platformlogger.RequestIDFromGin(c)),
		zap.Uint64("user_id", userID),
		zap.String("entity_type", "comment"),
		zap.Uint64("entity_id", comment.ID),
		zap.Uint64("post_id", req.PostID),
		zap.String("result", "success"),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)

	response.Success(c, comment)
}

// GetCommentList 获取评论列表
// @Summary 获取评论列表
// @Description 获取指定分享的评论列表
// @Tags 评论
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param post_id query int true "分享ID"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(20)
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/comments [get]
func (h *CommentHandler) GetCommentList(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "comment"),
		zap.String("operation", "get_comment_list"),
	)

	postIDStr := c.Query("post_id")
	if postIDStr == "" {
		log.Warn("validation_failed",
			zap.String("event", "validation_failed"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
		)
		response.FromError(c, apperror.New(apperror.CodeInvalidParams, apperror.KindValidation, "post_id is required", http.StatusBadRequest))
		return
	}
	postID, err := strconv.ParseUint(postIDStr, 10, 64)
	if err != nil {
		log.Warn("invalid_post_id",
			zap.String("event", "invalid_post_id"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeInvalidParams, apperror.KindValidation, "帖子ID无效", http.StatusBadRequest))
		return
	}

	log = log.With(zap.Uint64("post_id", postID))

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	comments, total, err := h.commentService.GetCommentsByPostID(ctx, postID, page, size)
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

	log.Info("comment_list_retrieved",
		zap.String("event", "comment_list_retrieved"),
		zap.Int64("total", total),
	)

	response.Success(c, gin.H{
		"list":  comments,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// GetCommentDetail 获取评论详情
// @Summary 获取评论详情
// @Description 获取指定ID的评论详情
// @Tags 评论
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "评论ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/comments/{id} [get]
func (h *CommentHandler) GetCommentDetail(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "comment"),
		zap.String("operation", "get_comment_detail"),
	)

	commentIDStr := c.Param("id")
	commentID, err := strconv.ParseUint(commentIDStr, 10, 64)
	if err != nil {
		log.Warn("invalid_comment_id",
			zap.String("event", "invalid_comment_id"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeInvalidParams, apperror.KindValidation, "评论ID无效", http.StatusBadRequest))
		return
	}

	log = log.With(zap.Uint64("comment_id", commentID))

	comment, err := h.commentService.GetCommentByID(ctx, commentID)
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
	if comment == nil {
		log.Warn("comment_not_found",
			zap.String("event", "comment_not_found"),
		)
		response.NotFound(c, "comment not found")
		return
	}

	log.Info("comment_detail_retrieved",
		zap.String("event", "comment_detail_retrieved"),
	)

	response.Success(c, comment)
}

type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required" example:"更新后的评论内容"`
}

// UpdateComment 更新评论
// @Summary 更新评论
// @Description 更新自己的评论内容
// @Tags 评论
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "评论ID"
// @Param body body UpdateCommentRequest true "评论内容"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/comments/{id} [put]
func (h *CommentHandler) UpdateComment(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "comment"),
		zap.String("operation", "update_comment"),
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

	commentIDStr := c.Param("id")
	commentID, err := strconv.ParseUint(commentIDStr, 10, 64)
	if err != nil {
		log.Warn("invalid_comment_id",
			zap.String("event", "invalid_comment_id"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeInvalidParams, apperror.KindValidation, "评论ID无效", http.StatusBadRequest))
		return
	}

	log = log.With(zap.Uint64("comment_id", commentID), zap.Uint64("user_id", userID))

	var req UpdateCommentRequest
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

	err = h.commentService.UpdateComment(ctx, commentID, userID, req.Content)
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

	log.Info("comment_updated",
		zap.String("event", "comment_updated"),
		zap.String("entity_type", "comment"),
		zap.Uint64("entity_id", commentID),
	)
	platformlogger.Audit(c.Request.Context(), "comment.update",
		zap.String("request_id", platformlogger.RequestIDFromGin(c)),
		zap.Uint64("user_id", userID),
		zap.String("entity_type", "comment"),
		zap.Uint64("entity_id", commentID),
		zap.String("result", "success"),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)

	response.Success(c, "update success")
}

// DeleteComment 删除评论
// @Summary 删除评论
// @Description 删除自己的评论
// @Tags 评论
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "评论ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/comments/{id} [delete]
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "comment"),
		zap.String("operation", "delete_comment"),
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

	commentIDStr := c.Param("id")
	commentID, err := strconv.ParseUint(commentIDStr, 10, 64)
	if err != nil {
		log.Warn("invalid_comment_id",
			zap.String("event", "invalid_comment_id"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeInvalidParams, apperror.KindValidation, "评论ID无效", http.StatusBadRequest))
		return
	}

	log = log.With(zap.Uint64("comment_id", commentID), zap.Uint64("user_id", userID))

	err = h.commentService.DeleteComment(ctx, commentID, userID)
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

	log.Info("comment_deleted",
		zap.String("event", "comment_deleted"),
		zap.String("entity_type", "comment"),
		zap.Uint64("entity_id", commentID),
	)
	platformlogger.Audit(c.Request.Context(), "comment.delete",
		zap.String("request_id", platformlogger.RequestIDFromGin(c)),
		zap.Uint64("user_id", userID),
		zap.String("entity_type", "comment"),
		zap.Uint64("entity_id", commentID),
		zap.String("result", "success"),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)

	response.Success(c, "delete success")
}
