package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trailmemo/internal/agent/dto"
	"github.com/trailmemo/pkg/apperror"
	"github.com/trailmemo/pkg/response"
)

// @Summary 一句话生成路线草稿
// @Description 自然语言→结构化路线草稿（打卡点/预算/时长）
// @Tags 路线
// @Accept json
// @Produce json
// @Param body body dto.RouteDraftRequest true "路线需求"
// @Success 200 {object} response.Response{data=dto.RouteDraftResponse}
// @Router /agent/routes/draft [post]
// 对应设计文档 §7 —— 一句话生成路线草稿。
func (h *Handler) CreateRouteDraft(c *gin.Context) {
	userID, ok := middlewareGetUserID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	var req dto.RouteDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误：query 为必填项")
		return
	}

	result, err := h.agentService.CreateRouteDraftWithSession(c.Request.Context(), userID, &req)
	if err != nil {
		response.FromError(c, apperror.New(apperror.CodeExternalError, apperror.KindExternal, err.Error(), http.StatusServiceUnavailable))
		return
	}

	response.Success(c, result)
}

// @Summary 提交产物到业务实体
// @Description 用户确认后将草稿导入为真实Route或Post，支持幂等
// @Tags 产物
// @Accept json
// @Produce json
// @Param artifact_id path string true "产物ID"
// @Param body body dto.ArtifactCommitRequest true "提交请求"
// @Success 200 {object} response.Response{data=dto.ArtifactCommitResponse}
// @Router /agent/artifacts/{artifact_id}/commit [post]
// 对应设计文档 §8 —— 用户确认后将草稿导入为真实业务实体。
func (h *Handler) CommitArtifact(c *gin.Context) {
	userID, ok := middlewareGetUserID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	artifactID := c.Param("artifact_id")
	if artifactID == "" {
		response.BadRequest(c, "artifact_id 不能为空")
		return
	}

	var req dto.ArtifactCommitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误：commit_type 和 idempotency_key 为必填项")
		return
	}

	result, err := h.agentService.CommitArtifact(c.Request.Context(), userID, artifactID, &req)
	if err != nil {
		response.FromError(c, apperror.New(apperror.CodeExternalError, apperror.KindExternal, err.Error(), http.StatusConflict))
		return
	}

	response.Success(c, result)
}

// ApproveArtifact 处理 POST /agent/artifacts/:artifact_id/approve 请求。
// 用户显式确认 AI 产物，后续 commit 才允许执行写操作。
func (h *Handler) ApproveArtifact(c *gin.Context) {
	userID, ok := middlewareGetUserID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	artifactID := c.Param("artifact_id")
	if artifactID == "" {
		response.BadRequest(c, "artifact_id 不能为空")
		return
	}

	result, err := h.agentService.ApproveArtifact(c.Request.Context(), userID, artifactID)
	if err != nil {
		response.FromError(c, apperror.New(apperror.CodeConflict, apperror.KindConflict, err.Error(), http.StatusConflict))
		return
	}

	response.Success(c, result)
}
