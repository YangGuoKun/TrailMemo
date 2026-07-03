package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/trailmemo/internal/agent/dto"
	"github.com/trailmemo/pkg/response"
)

// @Summary 会话列表
// @Description 获取当前用户的所有Agent对话会话，按时间倒序
// @Tags 会话
// @Produce json
// @Success 200 {object} response.Response{data=dto.SessionListResponse}
// @Router /agent/sessions [get]
func (h *Handler) ListSessions(c *gin.Context) {
	userID, ok := middlewareGetUserID(c)
	if !ok { response.Unauthorized(c, "未登录"); return }
	result, err := h.agentService.ListSessions(c.Request.Context(), userID)
	if err != nil { response.Fail(c, err.Error()); return }
	response.Success(c, result)
}

// @Summary 会话详情
// @Description 获取指定会话的消息历史，含expired标记
// @Tags 会话
// @Produce json
// @Param id path string true "会话ID"
// @Success 200 {object} response.Response{data=dto.SessionDetailResponse}
// @Router /agent/sessions/{id} [get]
func (h *Handler) GetSession(c *gin.Context) {
	userID, ok := middlewareGetUserID(c)
	if !ok { response.Unauthorized(c, "未登录"); return }
	sessionID := c.Param("id")
	result, err := h.agentService.GetSession(c.Request.Context(), userID, sessionID)
	if err != nil { response.Fail(c, err.Error()); return }
	response.Success(c, result)
}

// @Summary 删除会话
// @Description 同时清理 MySQL 元数据和 Redis 消息历史
// @Tags 会话
// @Produce json
// @Param id path string true "会话ID"
// @Success 200 {object} response.Response
// @Router /agent/sessions/{id} [delete]
func (h *Handler) DeleteSession(c *gin.Context) {
	userID, ok := middlewareGetUserID(c)
	if !ok { response.Unauthorized(c, "未登录"); return }
	sessionID := c.Param("id")
	if err := h.agentService.DeleteSession(c.Request.Context(), userID, sessionID); err != nil {
		response.Fail(c, err.Error()); return
	}
	response.Success(c, "会话已删除")
}

// @Summary 重命名会话
// @Description 修改会话标题
// @Tags 会话
// @Accept json
// @Produce json
// @Param id path string true "会话ID"
// @Param body body dto.RenameSessionRequest true "新标题"
// @Success 200 {object} response.Response
// @Router /agent/sessions/{id}/title [put]
func (h *Handler) RenameSession(c *gin.Context) {
	userID, ok := middlewareGetUserID(c)
	if !ok { response.Unauthorized(c, "未登录"); return }
	sessionID := c.Param("id")
	var req dto.RenameSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, "title 不能为空"); return }
	if err := h.agentService.RenameSession(c.Request.Context(), userID, sessionID, req.Title); err != nil {
		response.Fail(c, err.Error()); return
	}
	response.Success(c, "已重命名")
}
