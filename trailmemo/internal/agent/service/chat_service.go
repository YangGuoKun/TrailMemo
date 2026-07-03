package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/trailmemo/internal/agent/dto"
	"github.com/trailmemo/internal/agent/llm"
	"github.com/trailmemo/internal/agent/memory"
	agentruntime "github.com/trailmemo/internal/agent/runtime"
)

// ChatLoop 执行有界 Agent Loop 对话。
func (s *AgentService) ChatLoop(ctx context.Context, userID uint64, message, sessionID string) (*agentruntime.LoopResult, error) {
	if !s.cfg.Enabled {
		return nil, fmt.Errorf("agent 未启用")
	}

	router := agentruntime.NewIntentRouter()
	route := router.Route(message)

	const threshold = 0.25
	if route.Confidence >= threshold && route.Mode == agentruntime.ExecutionModeWorkflow {
		result, err := s.routeToWorkflow(ctx, userID, message, sessionID, route)
		if sessionID != "" && s.sessMem != nil && result != nil && result.Content != "" {
			_ = s.sessMem.AppendMessages(ctx, sessionID, []memory.SessionMessage{
				{Role: "user", Content: message},
				{Role: "assistant", Content: result.Content},
			}, 50)
			s.EnsureSession(ctx, userID, sessionID, message, "")
		}
		return result, err
	}
	return s.runAgentLoop(ctx, userID, message, sessionID)
}

func (s *AgentService) runAgentLoop(ctx context.Context, userID uint64, message, sessionID string) (*agentruntime.LoopResult, error) {
	// 1. 只校验用户原始输入
	if err := s.guardrail.CheckInput(ctx, message); err != nil {
		return &agentruntime.LoopResult{Content: "输入不符合要求，请重新描述", FinishReason: "error"}, nil
	}

	// 2. 从 Redis 加载历史 + 智能压缩
	var history []llm.Message
	if sessionID != "" && s.sessMem != nil {
		msgs, _ := s.sessMem.GetMessages(ctx, sessionID, 100)
		history = s.compressHistory(msgs)
	}

	// 3. 追加当前输入
	history = append(history, llm.Message{Role: "user", Content: message})

	lc := &agentruntime.LoopContext{
		Ctx: ctx, RunID: uuid.NewString(), UserID: userID, SessionID: sessionID,
		Messages: history, ToolRegistry: s.toolReg, LLMClient: s.llmClient,
		RunStore: s.runStore, SessionMem: s.sessMem,
		Config: agentruntime.DefaultLoopConfig(),
	}

	result, err := agentruntime.Run(lc)
	if err != nil && result.Content == "" {
		return result, err
	}

	if sessionID != "" && s.sessMem != nil && result.Content != "" {
		_ = s.sessMem.AppendMessages(ctx, sessionID, []memory.SessionMessage{
			{Role: "user", Content: message},
			{Role: "assistant", Content: result.Content},
		}, 50)
	}
	if sessionID != "" {
		s.EnsureSession(ctx, userID, sessionID, message, "")
	}
	fmt.Println("[Agent]", result.Content)
	return result, nil
}

// ChatStreamLoop 执行真正的端到端流式 Agent Loop 对话。
func (s *AgentService) ChatStreamLoop(ctx context.Context, userID uint64, message, sessionID string, ch chan<- string) error {
	defer close(ch)

	if !s.cfg.Enabled {
		ch <- "agent 未启用"
		return nil
	}

	router := agentruntime.NewIntentRouter()
	route := router.Route(message)
	const threshold = 0.25

	if route.Confidence >= threshold && route.Mode == agentruntime.ExecutionModeWorkflow {
		result, _ := s.routeToWorkflow(ctx, userID, message, sessionID, route)
		if result != nil {
			for _, r := range result.Content {
				ch <- string(r)
			}
		}
		if sessionID != "" && s.sessMem != nil && result != nil && result.Content != "" {
			_ = s.sessMem.AppendMessages(ctx, sessionID, []memory.SessionMessage{
				{Role: "user", Content: message},
				{Role: "assistant", Content: result.Content},
			}, 50)
			s.EnsureSession(ctx, userID, sessionID, message, "")
		}
		return nil
	}

	if err := s.guardrail.CheckInput(ctx, message); err != nil {
		ch <- `{"error":"输入不符合要求"}`
		return nil
	}
	var history []llm.Message
	if sessionID != "" && s.sessMem != nil {
		msgs, _ := s.sessMem.GetMessages(ctx, sessionID, 100)
		history = s.compressHistory(msgs)
	}
	history = append(history, llm.Message{Role: "user", Content: message})

	lc := &agentruntime.LoopContext{
		Ctx: ctx, RunID: uuid.NewString(), UserID: userID, SessionID: sessionID,
		Messages: history, ToolRegistry: s.toolReg, LLMClient: s.llmClient,
		RunStore: s.runStore, SessionMem: s.sessMem,
		Config: agentruntime.DefaultLoopConfig(),
	}

	result, _ := agentruntime.RunStream(lc, ch)

	if sessionID != "" && s.sessMem != nil && result != nil && result.Content != "" {
		_ = s.sessMem.AppendMessages(ctx, sessionID, []memory.SessionMessage{
			{Role: "user", Content: message},
			{Role: "assistant", Content: result.Content},
		}, 50)
	}
	if sessionID != "" {
		s.EnsureSession(ctx, userID, sessionID, message, "")
	}
	return nil
}

func (s *AgentService) routeToWorkflow(ctx context.Context, userID uint64, message, sessionID string, route agentruntime.RouteResult) (*agentruntime.LoopResult, error) {
	switch route.Intent {
	case agentruntime.IntentRouteDraft:
		return s.routeDraft(ctx, userID, message)
	case agentruntime.IntentRouteRemix:
		return s.routeRemix(ctx, userID, message)
	case agentruntime.IntentRecommend:
		return s.routeRecommend(ctx, userID, message)
	case agentruntime.IntentTravelNote:
		return s.routeTravelNote(ctx, userID, message)
	default:
		return &agentruntime.LoopResult{Content: "请更详细描述你的需求", Steps: 0, FinishReason: "stop"}, nil
	}
}

func (s *AgentService) routeDraft(ctx context.Context, userID uint64, message string) (*agentruntime.LoopResult, error) {
	req := &dto.RouteDraftRequest{Query: message, Days: 2}
	resp, err := s.CreateRouteDraft(ctx, userID, req)
	if err != nil {
		return &agentruntime.LoopResult{Content: "路线生成失败，请稍后再试", FinishReason: "error"}, nil
	}
	return &agentruntime.LoopResult{Content: fmt.Sprintf("已为你生成路线「%s」（%s→%s），共%d个打卡点，预估费用%.0f元。",
		resp.RouteDraft.Title, resp.RouteDraft.StartCity, resp.RouteDraft.EndCity, len(resp.RouteDraft.Checkpoints), resp.RouteDraft.EstimatedBudget), Steps: 1, FinishReason: "stop"}, nil
}

func (s *AgentService) routeRemix(ctx context.Context, userID uint64, message string) (*agentruntime.LoopResult, error) {
	req := &dto.RemixRequest{Query: message}
	resp, err := s.RemixRoute(ctx, userID, "0", req)
	if err != nil {
		return &agentruntime.LoopResult{Content: "改造失败：" + err.Error(), FinishReason: "error"}, nil
	}
	changes := ""
	for _, c := range resp.ChangeSummary {
		changes += fmt.Sprintf("· %s：%s（%s）\n", c.Action, c.Point, c.Reason)
	}
	return &agentruntime.LoopResult{Content: fmt.Sprintf("已生成改造路线「%s」\n变更说明：\n%s", resp.RouteDraft.Title, changes), Steps: 1, FinishReason: "stop"}, nil
}

func (s *AgentService) routeRecommend(ctx context.Context, userID uint64, message string) (*agentruntime.LoopResult, error) {
	req := &dto.RecommendRequest{Query: message}
	resp, err := s.Recommend(ctx, userID, req)
	if err != nil {
		return &agentruntime.LoopResult{Content: "推荐失败，请描述更具体的需求", FinishReason: "error"}, nil
	}
	result := "为你推荐以下目的地：\n"
	for i, item := range resp.Items {
		result += fmt.Sprintf("%d. %s（%s）%s 预估%d元/%d天\n", i+1, item.Title, item.City, item.Reason, item.EstimatedBudget, item.Days)
	}
	return &agentruntime.LoopResult{Content: result, Steps: 1, FinishReason: "stop"}, nil
}

func (s *AgentService) routeTravelNote(ctx context.Context, userID uint64, message string) (*agentruntime.LoopResult, error) {
	var routeID uint64
	fmt.Sscanf(message, "%d", &routeID)
	if routeID == 0 {
		return &agentruntime.LoopResult{Content: "请提供路线ID来生成游记，例如「为路线123生成游记」", Steps: 0, FinishReason: "stop"}, nil
	}
	req := &dto.TravelNoteRequest{RouteID: routeID, Style: "story", IncludeCheckinContent: true}
	resp, err := s.GenerateTravelNote(ctx, userID, req)
	if err != nil {
		return &agentruntime.LoopResult{Content: "游记生成失败：" + err.Error(), FinishReason: "error"}, nil
	}
	return &agentruntime.LoopResult{Content: fmt.Sprintf("已生成游记「%s」\n\n%s\n\n标签：%v", resp.Title, truncate(resp.Content, 200), resp.SuggestedTags), Steps: 1, FinishReason: "stop"}, nil
}

// compressHistory 智能压缩历史消息以适应 LLM token 预算。
// 策略：最近 6 条保留完整内容，更早的只保留首条用户消息作为上下文摘要。
func (s *AgentService) compressHistory(msgs []memory.SessionMessage) []llm.Message {
	const maxRecent = 6     // 保留最近 N 条
	const summaryMax = 2000 // 摘要最大字符数

	if len(msgs) <= maxRecent {
		result := make([]llm.Message, 0, len(msgs))
		for _, m := range msgs {
			result = append(result, llm.Message{Role: m.Role, Content: m.Content})
		}
		return result
	}

	// 早期消息 → 摘要
	earlyMsgs := msgs[:len(msgs)-maxRecent]
	summary := "【历史对话摘要】"
	for _, m := range earlyMsgs {
		content := m.Content
		if len([]rune(content)) > 80 {
			content = string([]rune(content)[:80]) + "..."
		}
		summary += fmt.Sprintf("\n%s: %s", m.Role, content)
	}
	if len([]rune(summary)) > summaryMax {
		summary = string([]rune(summary)[:summaryMax]) + "..."
	}

	result := []llm.Message{{Role: "system", Content: summary}}
	for _, m := range msgs[len(msgs)-maxRecent:] {
		result = append(result, llm.Message{Role: m.Role, Content: m.Content})
	}
	return result
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
