// Package runtime 实现 Agent 有界推理循环（Bounded Agent Loop）。
// 不同于 Workflow 的固定步骤，Loop 允许 LLM 在有限轮数内自主决定调用哪些工具。
// 对应设计文档 ADR-002：开放问答使用 bounded agent loop。
package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/trailmemo/internal/agent/llm"
	"github.com/trailmemo/internal/agent/memory"
	"github.com/trailmemo/internal/agent/tools"
	"github.com/trailmemo/internal/platform/logger"
	"go.uber.org/zap"
)

// LoopConfig 控制 Agent Loop 的行为边界。
type LoopConfig struct {
	MaxSteps     int           // 最大推理步数（默认6）
	MaxToolCalls int           // 单次运行最大工具调用次数（默认10）
	Timeout      time.Duration // 总超时
}

// DefaultLoopConfig 返回安全的默认循环配置。
func DefaultLoopConfig() LoopConfig {
	return LoopConfig{MaxSteps: 6, MaxToolCalls: 10, Timeout: 30 * time.Second}
}

// LoopContext 是一次 Agent Loop 执行的完整上下文。
type LoopContext struct {
	Ctx           context.Context
	RunID         string
	UserID        uint64
	SessionID     string
	Messages      []llm.Message         // 当前对话历史
	ToolRegistry  *tools.Registry       // 工具注册中心
	LLMClient     *llm.Client           // LLM 客户端
	RunStore      *memory.RunStore      // 运行记录
	SessionMem    *memory.SessionMemory // 会话记忆
	Config        LoopConfig
	ToolCallCount int // 已使用的工具调用次数
	TotalTokens   int // 累计 token
}

// LoopResult 是 Agent Loop 的最终输出。
type LoopResult struct {
	Content    string        `json:"content"`     // 最终回复
	Steps      int           `json:"steps"`        // 消耗的步数
	ToolCalls  int           `json:"tool_calls"`   // 工具调用次数
	TotalTokens int          `json:"total_tokens"` // 累计 token
	LatencyMs  int64         `json:"latency_ms"`   // 总耗时
	FinishReason string       `json:"finish_reason"` // stop / max_steps / timeout / error
}

// Run 执行有界 Agent Loop。
// 循环流程：发送消息 → LLM 决定调用工具或给出最终回复 → 执行工具 → 将结果反馈给 LLM → 重复。
func Run(lc *LoopContext) (*LoopResult, error) {
	start := time.Now()
	log := logger.FromContext(lc.Ctx).With(zap.String("run_id", lc.RunID))

	systemPrompt := llm.Message{Role: "system", Content: `你是迹忆旅图的AI旅行助手。你可以：
1. 回答旅行相关问题
2. 使用工具查询公开路线、打卡记录、社区帖子
3. 根据用户需求和偏好提供个性化建议

规则：
- 用中文回答
- 需要具体数据时调用工具
- 一次最多调2个工具
- 如果用户只是闲聊，直接回答不需要调工具`}

	messages := append([]llm.Message{systemPrompt}, lc.Messages...)
	toolDefs := convertToolDefs(lc.ToolRegistry)

	for step := 1; step <= lc.Config.MaxSteps; step++ {
		// 检查上下文超时
		select {
		case <-lc.Ctx.Done():
			return &LoopResult{Steps: step - 1, ToolCalls: lc.ToolCallCount, TotalTokens: lc.TotalTokens, LatencyMs: time.Since(start).Milliseconds(), FinishReason: "timeout"}, lc.Ctx.Err()
		default:
		}

		// LLM 推理
		resp, err := lc.LLMClient.Chat(lc.Ctx, llm.ChatRequest{
			Messages:   messages,
			MaxTokens:  2000,
			Tools:      toolDefs,
			ToolChoice: "auto",
		})
		if err != nil {
			return &LoopResult{Steps: step - 1, ToolCalls: lc.ToolCallCount, TotalTokens: lc.TotalTokens, LatencyMs: time.Since(start).Milliseconds(), FinishReason: "error", Content: "AI服务暂时不可用"}, err
		}
		lc.TotalTokens += resp.Usage.TotalTokens

		// 无工具调用 → LLM 给出了最终回复
		if len(resp.ToolCalls) == 0 {
			log.Info("agent_loop_completed", zap.Int("steps", step), zap.Int("total_tokens", lc.TotalTokens))
			return &LoopResult{Content: resp.Content, Steps: step, ToolCalls: lc.ToolCallCount, TotalTokens: lc.TotalTokens, LatencyMs: time.Since(start).Milliseconds(), FinishReason: "stop"}, nil
		}

		// 执行工具调用
		for _, tc := range resp.ToolCalls {
			if lc.ToolCallCount >= lc.Config.MaxToolCalls {
				messages = append(messages, llm.Message{Role: "assistant", Content: "已达到最大工具调用次数，请给出最终回答。"})
				break
			}

			toolResult, toolErr := lc.ToolRegistry.Execute(lc.Ctx, tc.Function.Name, json.RawMessage(tc.Function.Arguments))
			lc.ToolCallCount++

			resultContent := "工具执行成功"
			if toolErr != nil || !toolResult.Success {
				resultContent = fmt.Sprintf("工具执行失败: %v", toolErr)
			} else if toolResult.Data != nil {
				b, _ := json.Marshal(toolResult.Data)
				resultContent = string(b)
			}

			// 将工具结果追加到消息历史
			messages = append(messages,
				llm.Message{Role: "assistant", Content: "", ToolCalls: []llm.ToolCall{tc}},
				llm.Message{Role: "tool", Content: resultContent, ToolCallID: tc.ID},
			)
			log.Info("agent_tool_executed", zap.String("tool", tc.Function.Name), zap.Int("step", step))
		}
	}

	return &LoopResult{Content: "抱歉，处理超时了，请重新描述你的需求。", Steps: lc.Config.MaxSteps, ToolCalls: lc.ToolCallCount, TotalTokens: lc.TotalTokens, LatencyMs: time.Since(start).Milliseconds(), FinishReason: "max_steps"}, nil
}

// RunStream 执行流式有界 Agent Loop——推送进度事件和最终 token。
// ch 由调用方创建，RunStream 不关闭它。
// 每个 ch 消息都是完整 JSON：{"progress":"..."} / {"content":"c"} / {"error":"..."}
func RunStream(lc *LoopContext, ch chan<- string) (*LoopResult, error) {
	start := time.Now()
	log := logger.FromContext(lc.Ctx).With(zap.String("run_id", lc.RunID))

	systemPrompt := llm.Message{Role: "system", Content: "你是迹忆旅图的AI旅行助手。用中文回答。需要具体数据时调用工具。"}
	messages := append([]llm.Message{systemPrompt}, lc.Messages...)
	toolDefs := convertToolDefs(lc.ToolRegistry)

	// 安全推送 JSON 的辅助函数
	push := func(format string, args ...interface{}) {
		select {
		case ch <- fmt.Sprintf(format, args...):
		default:
		}
	}
	pushJSON := func(kv ...string) {
		parts := make([]string, 0, len(kv)/2)
		for i := 0; i < len(kv); i += 2 {
			b, _ := json.Marshal(kv[i+1])
			parts = append(parts, fmt.Sprintf(`"%s":%s`, kv[i], string(b)))
		}
		push("{%s}", stringsJoin(parts, ","))
	}

	pushJSON("progress", "正在分析需求...")

	for step := 1; step <= lc.Config.MaxSteps; step++ {
		select {
		case <-lc.Ctx.Done():
			pushJSON("error", "连接超时")
			return &LoopResult{Steps: step - 1, ToolCalls: lc.ToolCallCount, LatencyMs: time.Since(start).Milliseconds(), FinishReason: "timeout"}, lc.Ctx.Err()
		default:
		}

		// 中间步骤：非流式判断是否需要工具
		resp, err := lc.LLMClient.Chat(lc.Ctx, llm.ChatRequest{Messages: messages, MaxTokens: 2000, Tools: toolDefs, ToolChoice: "auto"})
		if err != nil {
			pushJSON("error", "AI服务暂不可用")
			return &LoopResult{Steps: step, ToolCalls: lc.ToolCallCount, LatencyMs: time.Since(start).Milliseconds(), FinishReason: "error"}, err
		}
		lc.TotalTokens += resp.Usage.TotalTokens

		// 无工具调用 → 流式输出最终回复
		if len(resp.ToolCalls) == 0 {
			streamCh := make(chan llm.StreamChunk, 64)
			go lc.LLMClient.StreamToChannel(lc.Ctx, llm.ChatRequest{Messages: messages, MaxTokens: 2000}, streamCh)
			for chunk := range streamCh {
				if chunk.Error != nil || chunk.Done { break }
				if chunk.Content != "" { pushJSON("content", chunk.Content) }
			}
			log.Info("agent_stream_completed", zap.Int("steps", step))
			return &LoopResult{Content: resp.Content, Steps: step, ToolCalls: lc.ToolCallCount, TotalTokens: lc.TotalTokens, LatencyMs: time.Since(start).Milliseconds(), FinishReason: "stop"}, nil
		}

		// 有工具调用 → 推送进度 + 执行
		toolNames := make([]string, 0, len(resp.ToolCalls))
		for _, tc := range resp.ToolCalls { toolNames = append(toolNames, tc.Function.Name) }
		pushJSON("progress", "正在查询："+stringsJoin(toolNames, "、"))

		for _, tc := range resp.ToolCalls {
			if lc.ToolCallCount >= lc.Config.MaxToolCalls { break }
			toolResult, toolErr := lc.ToolRegistry.Execute(lc.Ctx, tc.Function.Name, json.RawMessage(tc.Function.Arguments))
			lc.ToolCallCount++
			resultContent := "工具执行成功"
			if toolErr != nil || !toolResult.Success {
				resultContent = fmt.Sprintf("工具执行失败: %v", toolErr)
			} else if toolResult.Data != nil {
				b, _ := json.Marshal(toolResult.Data)
				resultContent = string(b)
			}
			messages = append(messages,
				llm.Message{Role: "assistant", Content: "", ToolCalls: []llm.ToolCall{tc}},
				llm.Message{Role: "tool", Content: resultContent, ToolCallID: tc.ID})
			log.Info("agent_tool_executed", zap.String("tool", tc.Function.Name), zap.Int("step", step))
		}
	}

	pushJSON("error", "处理超时，请重试")
	return &LoopResult{Steps: lc.Config.MaxSteps, ToolCalls: lc.ToolCallCount, LatencyMs: time.Since(start).Milliseconds(), FinishReason: "max_steps"}, nil
}

func stringsJoin(parts []string, sep string) string {
	if len(parts) == 0 { return "" }
	result := parts[0]
	for i := 1; i < len(parts); i++ { result += sep + parts[i] }
	return result
}

// convertToolDefs 将 tools.Registry 中的工具转为 LLM 可用的 ToolDef 列表。
func convertToolDefs(reg *tools.Registry) []llm.ToolDef {
	if reg == nil { return nil }
	toolDefs := reg.GetToolDefs()
	defs := make([]llm.ToolDef, 0, len(toolDefs))
	for _, d := range toolDefs {
		defs = append(defs, llm.ToolDef{
			Function: llm.FunctionDef{
				Name:        d.Function.Name,
				Description: d.Function.Description,
				Parameters:  d.Function.Parameters,
			},
		})
	}
	return defs
}
