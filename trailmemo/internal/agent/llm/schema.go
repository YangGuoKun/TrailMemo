package llm

import (
	"encoding/json"
	"time"
)

// ── Request types ──────────────────────────────────
// 该结构体定义了OpenAI的ChatCompletion API的请求参数。
// 每个消息都有一个角色（role），一个内容（content），以及可能的工具调用（tool_calls）和工具调用ID（tool_call_id）。
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// 该结构体定义了OpenAI的ChatCompletion API的工具定义。
// 每个工具都有一个名称（name），一个描述（description），以及参数（parameters）。
type ToolDef struct {
	Function FunctionDef `json:"function"`
}

// 该结构体定义了OpenAI的ChatCompletion API的函数定义。
// 每个函数都有一个名称（name），一个描述（description），以及参数（parameters）。
type FunctionDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

// 该结构体定义了OpenAI的ChatCompletion API的工具调用。
// 每个工具调用都有一个ID（id），一个类型（type），以及函数调用（function）。
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// 该结构体定义了OpenAI的ChatCompletion API的函数调用。
// 每个函数调用都有一个名称（name），以及参数（arguments）。
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// 该结构体定义了OpenAI的ChatCompletion API的JSON格式响应。
// 每个响应都有一个类型（type），例如"json_object"。
type JSONFormat struct {
	Type string `json:"type"`
}

// 该结构体定义了OpenAI的ChatCompletion API的请求参数。
// 每个请求都有一个消息列表（messages），以及可能的最大令牌数（max_tokens）和工具定义（tools）。
type ChatRequest struct {
	Messages       []Message   `json:"messages"`
	MaxTokens      int         `json:"max_tokens,omitempty"`
	Temperature    float64     `json:"temperature,omitempty"`
	Tools          []ToolDef   `json:"tools,omitempty"`
	ToolChoice     string      `json:"tool_choice,omitempty"` // "auto", "none", or specific function
	ResponseFormat *JSONFormat `json:"response_format,omitempty"`
}

// ── Response types ──────────────────────────────────
// 该结构体定义了OpenAI的ChatCompletion API的响应参数。
// 每个响应都有一个内容（content），以及可能的工具调用（tool_calls）和使用统计（usage）。
// 同时，还包含了响应的延迟（latency）和错误（error）。
type ChatResponse struct {
	Content   string        `json:"content"`
	ToolCalls []ToolCall    `json:"tool_calls,omitempty"`
	Usage     Usage         `json:"usage"`
	Latency   time.Duration `json:"latency"`
	Error     error         `json:"-"`
}

// 该结构体定义了OpenAI的ChatCompletion API的使用统计。
// 每个使用统计都有一个提示令牌数（prompt_tokens），一个完成令牌数（completion_tokens），以及一个总令牌数（total_tokens）。
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ── Stream ──────────────────────────────────────────

// StreamChunk is a single token delta from a streaming response.
type StreamChunk struct {
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	Done      bool       `json:"done"`
	Error     error      `json:"-"`
}
