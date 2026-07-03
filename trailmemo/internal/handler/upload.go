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

type UploadHandler struct {
	uploadService service.UploadService
	userService   service.UserService
}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{
		uploadService: service.NewUploadService(),
		userService:   service.NewUserService(),
	}
}

func (h *UploadHandler) RegisterRoutes(r *gin.RouterGroup) {
	uploadGroup := r.Group("/upload")
	uploadGroup.Use(middleware.JWTAuth())
	{
		uploadGroup.POST("/avatar", h.UploadAvatar)
	}
}

// UploadAvatar 上传头像
// @Summary 上传用户头像
// @Description 上传图片文件作为用户头像
// @Tags 文件上传
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "头像图片文件"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/upload/avatar [post]
func (h *UploadHandler) UploadAvatar(c *gin.Context) {
	ctx := c.Request.Context()
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "upload"),
		zap.String("operation", "upload_avatar"),
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
		response.FromError(c, apperror.Wrap(err, apperror.CodeInternalError, apperror.KindInternal, "用户ID解析失败", http.StatusInternalServerError))
		return
	}

	log = log.With(zap.Uint64("user_id", userID))

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		log.Warn("validation_failed",
			zap.String("event", "validation_failed"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
			zap.String("filename", header.Filename),
			zap.Error(err),
		)
		response.FromError(c, apperror.Wrap(err, apperror.CodeInvalidParams, apperror.KindValidation, "文件是必需的", http.StatusBadRequest))
		return
	}
	defer file.Close()

	log = log.With(zap.String("filename", header.Filename))

	fileURL, err := h.uploadService.UploadAvatar(file, header.Filename)
	if err != nil {
		log.Error("upload_failed",
			zap.String("event", "upload_failed"),
			zap.String("error_kind", apperror.KindExternal),
			zap.Error(err),
		)
		response.FromError(c, err)
		return
	}

	log = log.With(zap.String("file_url", fileURL))

	if err := h.userService.UpdateUserInfo(ctx, userID, "", fileURL); err != nil {
		log.Error("update_user_failed",
			zap.String("event", "update_user_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		response.FromError(c, err)
		return
	}

	log.Info("avatar_uploaded",
		zap.String("event", "avatar_uploaded"),
		zap.Uint64("user_id", userID),
		zap.String("file_url", fileURL),
	)
	platformlogger.Audit(c.Request.Context(), "upload.avatar",
		zap.String("request_id", platformlogger.RequestIDFromGin(c)),
		zap.Uint64("user_id", userID),
		zap.String("action", "upload_avatar"),
		zap.String("file_url", fileURL),
		zap.String("result", "success"),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)

	response.Success(c, gin.H{
		"url": fileURL,
	})
}
