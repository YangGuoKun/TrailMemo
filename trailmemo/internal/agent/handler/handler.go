package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trailmemo/internal/agent/dto"
	agentservice "github.com/trailmemo/internal/agent/service"
	"github.com/trailmemo/internal/middleware"
	"github.com/trailmemo/pkg/apperror"
	"github.com/trailmemo/pkg/response"
)

type Handler struct {
	agentService *agentservice.AgentService
}

func NewHandler() *Handler { return &Handler{agentService: agentservice.NewAgentService()} }

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// 公开接口
	publicGroup := r.Group("/agent")
	{
		publicGroup.GET("/health", h.Health)
		publicGroup.GET("/capabilities", h.Capabilities)
		publicGroup.GET("/metrics", h.Metrics)
	}

	// 需登录接口
	ag := r.Group("/agent")
	ag.Use(middleware.JWTAuth())
	{
		ag.POST("/chat", h.Chat)
		ag.POST("/chat/stream", h.ChatStream)
		ag.POST("/recommend", h.Recommend)
		ag.POST("/routes/draft", h.CreateRouteDraft)
		ag.POST("/routes/:id/remix", h.RemixRoute)
		ag.POST("/notes/generate", h.GenerateTravelNote)
		ag.POST("/artifacts/:artifact_id/approve", h.ApproveArtifact)
		ag.POST("/artifacts/:artifact_id/commit", h.CommitArtifact)
		ag.GET("/preferences", h.GetPreferences)
		ag.PUT("/preferences", h.UpdatePreferences)
		ag.DELETE("/preferences/memory", h.DeleteMemory)
		ag.GET("/runs/:run_id", h.GetRunDetail)
		ag.GET("/sessions", h.ListSessions)
		ag.GET("/sessions/:id", h.GetSession)
		ag.DELETE("/sessions/:id", h.DeleteSession)
		ag.PUT("/sessions/:id/title", h.RenameSession)
	}
}

func middlewareGetUserID(c *gin.Context) (uint64, bool) {
	s, ok := c.Get("user_id")
	if !ok { return 0, false }
	uid, err := strconv.ParseUint(s.(string), 10, 64)
	return uid, err == nil
}

// @Summary Agent 健康检查
// @Description 返回 Agent 模块运行状态和 LLM 配置
// @Tags Agent
// @Produce json
// @Success 200 {object} response.Response{data=dto.HealthResponse}
// @Router /agent/health [get]
func (h *Handler) Health(c *gin.Context) { response.Success(c, h.agentService.Health()) }

// @Summary Agent 能力列表
// @Description 返回支持的意图、工具和运行模式
// @Tags Agent
// @Produce json
// @Success 200 {object} response.Response{data=dto.CapabilityResponse}
// @Router /agent/capabilities [get]
func (h *Handler) Capabilities(c *gin.Context) { response.Success(c, h.agentService.Capabilities()) }

// @Summary Agent 运行指标
// @Description 返回累计请求数、成功率、token用量、延迟分布
// @Tags Agent
// @Produce json
// @Success 200 {object} response.Response
// @Router /agent/metrics [get]
func (h *Handler) Metrics(c *gin.Context) { response.Success(c, h.agentService.GetMetrics()) }

// @Summary Agent 对话（非流式）
// @Description 有界推理循环处理用户消息，意图自动路由
// @Tags 对话
// @Accept json
// @Produce json
// @Param body body dto.ChatRequest true "对话请求"
// @Success 200 {object} response.Response{data=dto.ChatLoopResponse}
// @Router /agent/chat [post]
func (h *Handler) Chat(c *gin.Context) {
	userID, _ := middlewareGetUserID(c)
	var req dto.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, "message 为必填项"); return }

	result, err := h.agentService.ChatLoop(c.Request.Context(), userID, req.Message, req.SessionID)
	if err != nil {
		response.FromError(c, apperror.New(apperror.CodeExternalError, apperror.KindExternal, "AI服务暂时不可用", http.StatusServiceUnavailable))
		return
	}
	response.Success(c, dto.ChatLoopResponse{Content: result.Content, Steps: result.Steps, ToolCalls: result.ToolCalls, TotalTokens: result.TotalTokens, LatencyMs: result.LatencyMs, FinishReason: result.FinishReason})
}

// @Summary Agent 流式对话（SSE）
// @Description 通过 Server-Sent Events 实时推送 AI 回复 token
// @Tags 对话
// @Accept json
// @Produce text/event-stream
// @Param body body dto.ChatRequest true "对话请求"
// @Router /agent/chat/stream [post]
func (h *Handler) ChatStream(c *gin.Context) {
	userID, _ := middlewareGetUserID(c)
	var req dto.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, "message 为必填项"); return }

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	ch := make(chan string, 32)
	errCh := make(chan error, 1)
	go func() { errCh <- h.agentService.ChatStreamLoop(c.Request.Context(), userID, req.Message, req.SessionID, ch) }()

	c.Stream(func(w io.Writer) bool {
		select {
		case evt, ok := <-ch:
			if !ok { return false }
			fmt.Fprintf(w, "%s\n", evt) // 每行一个 JSON 事件
			return true
		case err := <-errCh:
			if err != nil { fmt.Fprintf(w, "{\"error\":%q}\n", err.Error()) }
			return false
		case <-c.Request.Context().Done():
			return false
		}
	})
}
