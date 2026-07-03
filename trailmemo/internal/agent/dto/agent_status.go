package dto

import "github.com/trailmemo/internal/agent/tools"

type CapabilityResponse struct {
	Enabled      bool                   `json:"enabled"`
	DefaultMode  string                 `json:"default_mode"`
	MaxSteps     int                    `json:"max_steps"`
	MaxToolCalls int                    `json:"max_tool_calls"`
	Intents      []string               `json:"intents"`
	Tools        []tools.ToolDescriptor `json:"tools"`
	Stage        string                 `json:"stage"`
}

type HealthResponse struct {
	Status         string `json:"status"`
	Enabled        bool   `json:"enabled"`
	Stage          string `json:"stage"`
	LLMConfigured  bool   `json:"llm_configured"`
	DefaultMode    string `json:"default_mode"`
	RequestTimeout string `json:"request_timeout"`
	StreamTimeout  string `json:"stream_timeout"`
}

type ChatResponse struct {
	Content   string   `json:"content"`
	Usage     UsageInfo `json:"usage"`
	LatencyMs int64    `json:"latency_ms"`
}

type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatRequest struct {
	Message   string `json:"message" binding:"required"`
	SessionID string `json:"session_id"` // 可选，用于多轮对话
}

type ChatLoopResponse struct {
	Content      string `json:"content"`
	Steps        int    `json:"steps"`
	ToolCalls    int    `json:"tool_calls"`
	TotalTokens  int    `json:"total_tokens"`
	LatencyMs    int64  `json:"latency_ms"`
	FinishReason string `json:"finish_reason"`
}
