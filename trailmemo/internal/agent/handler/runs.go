package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trailmemo/pkg/apperror"
	"github.com/trailmemo/pkg/response"
)

// GetRunDetail 处理 GET /agent/runs/:run_id 请求。
// 返回一次 Agent 运行的主记录、步骤和产物。
func (h *Handler) GetRunDetail(c *gin.Context) {
	userID, ok := middlewareGetUserID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	runID := c.Param("run_id")
	if runID == "" {
		response.BadRequest(c, "run_id 不能为空")
		return
	}

	result, err := h.agentService.GetRunDetail(c.Request.Context(), userID, runID)
	if err != nil {
		response.FromError(c, apperror.New(apperror.CodeResourceNotFound, apperror.KindNotFound, "运行记录不存在", http.StatusNotFound))
		return
	}

	response.Success(c, result)
}
