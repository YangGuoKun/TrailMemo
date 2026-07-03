// Package workflow 实现 Agent 工作流编排。
// 严格按照 AI_AGENT_PHASE3_WORKFLOW_DESIGN.md 设计。
// 每个 Workflow 是固定步骤序列，LLM 只生成结构化产物，不直接写业务库。
package workflow

import (
	"context"

	"github.com/trailmemo/internal/agent/guardrail"
	"github.com/trailmemo/internal/agent/llm"
	"github.com/trailmemo/internal/agent/memory"
	"github.com/trailmemo/internal/agent/prompt"
	"github.com/trailmemo/internal/agent/runtime"
	"github.com/trailmemo/internal/agent/tools"
	"go.uber.org/zap"
)

type LLMChatClient interface {
	Chat(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error)
}

type RunRecorder interface {
	CreateRun(ctx context.Context, run *memory.AgentRun) error
	CompleteRun(ctx context.Context, runID string, totalTokens int, latencyMs int64) error
	FailRun(ctx context.Context, runID string, errorCode string, errorMsg string) error
	AddStep(ctx context.Context, step *memory.AgentStep) error
}

type ArtifactRecorder interface {
	SaveArtifact(ctx context.Context, artifact *memory.AgentArtifact) error
}

// WorkflowContext 是一次 Workflow 执行的完整上下文。
// 贯穿 handler → service → workflow → tool → LLM → memory。
type WorkflowContext struct {
	Ctx           context.Context       `json:"-"`           // Go context，含 request_id
	RequestID     string                `json:"request_id"`  // 请求追踪 ID
	RunID         string                `json:"run_id"`      // Agent 运行 ID
	UserID        uint64                `json:"user_id"`     // 用户 ID
	SessionID     string                `json:"session_id"`  // 会话 ID
	Intent        runtime.Intent        `json:"intent"`      // 意图类型
	Mode          runtime.ExecutionMode `json:"mode"`        // 执行模式
	Input         any                   `json:"input"`       // 原始请求输入
	LLMClient     LLMChatClient         `json:"-"`           // LLM 客户端
	PromptMgr     *prompt.Manager       `json:"-"`           // Prompt 管理器
	ToolRegistry  *tools.Registry       `json:"-"`           // 工具注册中心
	RunStore      RunRecorder           `json:"-"`           // 运行记录存储
	ArtifactStore ArtifactRecorder      `json:"-"`           // 产物存储
	SessionMem    *memory.SessionMemory `json:"-"`           // 会话记忆
	Guardrail     *guardrail.Service    `json:"-"`           // 安全护栏
	Logger        *zap.Logger           `json:"-"`           // 带 request_id 的 logger
	Preferences   *UserPrefs            `json:"preferences"` // 用户偏好快照
}

// Workflow 接口定义了可编排的工作流。
type Workflow interface {
	// Name 返回工作流名称。
	Name() string
	// Run 执行工作流，返回结构化产物。
	Run(wc *WorkflowContext) (*WorkflowResult, error)
}

// WorkflowResult 是工作流执行的统一返回结构。
type WorkflowResult struct {
	Artifact     any      `json:"artifact"`          // 生成的结构化产物
	ArtifactType string   `json:"artifact_type"`     // 产物类型
	ArtifactID   string   `json:"artifact_id"`       // 保存后的产物 ID
	TotalTokens  int      `json:"total_tokens"`      // token 总量
	LatencyMs    int64    `json:"latency_ms"`        // 总耗时
	Warnings     []string `json:"warnings"`          // 非致命警告
	NextAction   string   `json:"next_action"`       // 后续操作建议
	Approval     bool     `json:"approval_required"` // 是否需要用户确认
}

// UserPrefs 是用户偏好快照。
type UserPrefs struct {
	TravelStyles []string `json:"travel_styles"`
	BudgetLevel  string   `json:"budget_level"`
	Interests    []string `json:"interests"`
	Pace         string   `json:"pace"`
}

func (p *UserPrefs) IsEmpty() bool {
	return p == nil || (p.BudgetLevel == "" && p.Pace == "" && len(p.Interests) == 0)
}

// 原始 UserPrefs 定义（保留原有字段兼容）
type _UserPrefsLegacy struct {
	TravelStyles []string `json:"travel_styles"` // 旅行风格
	BudgetLevel  string   `json:"budget_level"`  // 预算等级
	Interests    []string `json:"interests"`     // 兴趣标签
	Pace         string   `json:"pace"`          // 节奏偏好
}
