package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trailmemo/internal/agent/dto"
	"github.com/trailmemo/pkg/apperror"
	"github.com/trailmemo/pkg/response"
)

// GenerateTravelNote 处理 POST /agent/notes/generate 请求。
// 从路线打卡记录生成游记草稿。
// @Summary 打卡生成游记
// @Description 从路线打卡记录自动生成游记草稿
// @Tags 游记
// @Accept json
// @Produce json
// @Param body body dto.TravelNoteRequest true "游记请求"
// @Success 200 {object} response.Response{data=dto.TravelNoteResponse}
// @Router /agent/notes/generate [post]
func (h *Handler) GenerateTravelNote(c *gin.Context) {
	userID, ok := middlewareGetUserID(c)
	if !ok { response.Unauthorized(c, "未登录"); return }

	var req dto.TravelNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误：route_id 为必填项")
		return
	}

	result, err := h.agentService.GenerateTravelNote(c.Request.Context(), userID, &req)
	if err != nil {
		response.FromError(c, apperror.New(apperror.CodeExternalError, apperror.KindExternal, err.Error(), http.StatusServiceUnavailable))
		return
	}

	response.Success(c, result)
}

// Recommend 处理 POST /agent/recommend 请求。
// 根据用户需求和偏好生成旅行推荐。
// @Summary 旅行推荐
// @Description 根据用户需求和历史偏好生成目的地推荐
// @Tags 推荐
// @Accept json
// @Produce json
// @Param body body dto.RecommendRequest true "推荐请求"
// @Success 200 {object} response.Response{data=dto.RecommendResponse}
// @Router /agent/recommend [post]
func (h *Handler) Recommend(c *gin.Context) {
	userID, ok := middlewareGetUserID(c)
	if !ok { response.Unauthorized(c, "未登录"); return }

	var req dto.RecommendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误：query 为必填项")
		return
	}

	result, err := h.agentService.Recommend(c.Request.Context(), userID, &req)
	if err != nil {
		response.FromError(c, apperror.New(apperror.CodeExternalError, apperror.KindExternal, err.Error(), http.StatusServiceUnavailable))
		return
	}

	response.Success(c, result)
}

// RemixRoute 处理 POST /agent/routes/:id/remix 请求。
// 基于公开路线按用户需求改造生成新路线草稿。
// @Summary 改造公开路线
// @Description 基于公开路线按需求改造（亲子版/情侣版/美食版）
// @Tags 路线
// @Accept json
// @Produce json
// @Param id path int true "原路线ID"
// @Param body body dto.RemixRequest true "改造需求"
// @Success 200 {object} response.Response{data=dto.RemixResponse}
// @Router /agent/routes/{id}/remix [post]
func (h *Handler) RemixRoute(c *gin.Context) {
	userID, ok := middlewareGetUserID(c)
	if !ok { response.Unauthorized(c, "未登录"); return }

	routeIDStr := c.Param("id")
	if routeIDStr == "" { response.BadRequest(c, "路线 ID 不能为空"); return }

	var req dto.RemixRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误：query 为必填项")
		return
	}

	result, err := h.agentService.RemixRoute(c.Request.Context(), userID, routeIDStr, &req)
	if err != nil {
		response.FromError(c, apperror.New(apperror.CodeExternalError, apperror.KindExternal, err.Error(), http.StatusServiceUnavailable))
		return
	}

	response.Success(c, result)
}
