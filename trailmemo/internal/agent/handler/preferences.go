package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/trailmemo/internal/agent/memory"
	"github.com/trailmemo/pkg/response"
)

// GetPreferences 获取当前用户的 AI 画像。
// @Summary 查看AI画像
// @Description 返回AI从用户行为中学习到的偏好快照
// @Tags 偏好
// @Produce json
// @Success 200 {object} response.Response
// @Router /agent/preferences [get]
func (h *Handler) GetPreferences(c *gin.Context) {
	userID, ok := middlewareGetUserID(c)
	if !ok { response.Unauthorized(c, "未登录"); return }
	result := h.agentService.GetPreferences(c.Request.Context(), userID)
	response.Success(c, result)
}

// UpdatePreferences 手动更新 AI 偏好。
// @Summary 手动更新偏好
// @Description 用户显式设置旅行偏好
// @Tags 偏好
// @Accept json
// @Produce json
// @Param body body memory.PreferenceUpdate true "偏好更新"
// @Success 200 {object} response.Response
// @Router /agent/preferences [put]
func (h *Handler) UpdatePreferences(c *gin.Context) {
	userID, ok := middlewareGetUserID(c)
	if !ok { response.Unauthorized(c, "未登录"); return }
	var req memory.PreferenceUpdate
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, "参数错误"); return }
	if err := h.agentService.UpdatePreferences(c.Request.Context(), userID, &req); err != nil {
		response.Fail(c, err.Error()); return
	}
	response.Success(c, "偏好已更新")
}

// DeleteMemory 清空 AI 记忆。
// @Summary 清空AI记忆
// @Description 删除所有AI学到的用户偏好数据
// @Tags 偏好
// @Produce json
// @Success 200 {object} response.Response
// @Router /agent/preferences/memory [delete]
func (h *Handler) DeleteMemory(c *gin.Context) {
	userID, ok := middlewareGetUserID(c)
	if !ok { response.Unauthorized(c, "未登录"); return }
	if err := h.agentService.DeleteMemory(c.Request.Context(), userID); err != nil {
		response.Fail(c, err.Error()); return
	}
	response.Success(c, "AI 记忆已清空")
}
