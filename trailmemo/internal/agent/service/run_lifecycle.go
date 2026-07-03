package service

import (
	"context"
	"fmt"
	"reflect"

	"github.com/trailmemo/internal/agent/memory"
	agentruntime "github.com/trailmemo/internal/agent/runtime"
	"github.com/trailmemo/internal/agent/workflow"
)

func (s *AgentService) startWorkflowRun(ctx context.Context, wc *workflow.WorkflowContext) error {
	if wc == nil || wc.RunStore == nil {
		return fmt.Errorf("workflow context or run store is nil")
	}
	return wc.RunStore.CreateRun(ctx, s.newAgentRunFromWorkflowContext(wc))
}

func (s *AgentService) completeWorkflowRun(ctx context.Context, wc *workflow.WorkflowContext, result *workflow.WorkflowResult) {
	if wc == nil || wc.RunStore == nil || result == nil {
		return
	}
	_ = wc.RunStore.CompleteRun(ctx, wc.RunID, result.TotalTokens, result.LatencyMs)
}

func (s *AgentService) failWorkflowRun(ctx context.Context, wc *workflow.WorkflowContext, code string, err error) {
	if wc == nil || wc.RunStore == nil {
		return
	}
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	_ = wc.RunStore.FailRun(ctx, wc.RunID, code, msg)
}

func (s *AgentService) newAgentRunFromWorkflowContext(wc *workflow.WorkflowContext) *memory.AgentRun {
	return &memory.AgentRun{
		RunID:        wc.RunID,
		UserID:       wc.UserID,
		SessionID:    wc.SessionID,
		Intent:       string(wc.Intent),
		Mode:         string(wc.Mode),
		Status:       string(agentruntime.RunStatusRunning),
		InputSummary: summarizeWorkflowInput(wc.Input),
		Model:        s.cfg.LLM.Model,
		PromptVer:    workflowPromptVersion(wc.Intent),
	}
}

func summarizeWorkflowInput(input any) string {
	if input == nil {
		return ""
	}
	value := reflect.Indirect(reflect.ValueOf(input))
	if !value.IsValid() || value.Kind() != reflect.Struct {
		return fmt.Sprintf("%v", input)
	}
	for _, field := range []string{"Query", "Message", "Title"} {
		f := value.FieldByName(field)
		if f.IsValid() && f.Kind() == reflect.String {
			return truncateRunSummary(f.String(), 200)
		}
	}
	if f := value.FieldByName("RouteID"); f.IsValid() && f.CanInterface() {
		return fmt.Sprintf("route_id=%v", f.Interface())
	}
	return truncateRunSummary(fmt.Sprintf("%v", input), 200)
}

func workflowPromptVersion(intent agentruntime.Intent) string {
	switch intent {
	case agentruntime.IntentRouteDraft:
		return "route_draft:v1"
	case agentruntime.IntentRecommend:
		return "recommend:v1"
	case agentruntime.IntentRouteRemix:
		return "route_remix:v1"
	case agentruntime.IntentTravelNote:
		return "travel_note:v1"
	default:
		return string(intent) + ":v1"
	}
}

func truncateRunSummary(text string, maxRunes int) string {
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}
	return string(runes[:maxRunes]) + "..."
}
