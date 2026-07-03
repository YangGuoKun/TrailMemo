package service

import (
	"context"
	"testing"

	"github.com/trailmemo/internal/agent/memory"
	agentruntime "github.com/trailmemo/internal/agent/runtime"
	"github.com/trailmemo/internal/agent/workflow"
	"github.com/trailmemo/internal/config"
)

func TestNewAgentRunFromWorkflowContext(t *testing.T) {
	svc := &AgentService{cfg: config.AgentConfig{LLM: config.LLMConfig{Model: "qwen-plus"}}}
	wc := &workflow.WorkflowContext{
		Ctx:       context.Background(),
		RunID:     "run-123",
		UserID:    42,
		SessionID: "session-abc",
		Intent:    agentruntime.IntentRouteDraft,
		Mode:      agentruntime.ExecutionModeWorkflow,
		Input:     &workflow.RouteDraftRequest{Query: "杭州两日游，预算1500，喜欢美食"},
	}

	run := svc.newAgentRunFromWorkflowContext(wc)

	if run.RunID != "run-123" {
		t.Fatalf("expected run id run-123, got %s", run.RunID)
	}
	if run.UserID != 42 {
		t.Fatalf("expected user id 42, got %d", run.UserID)
	}
	if run.SessionID != "session-abc" {
		t.Fatalf("expected session id session-abc, got %s", run.SessionID)
	}
	if run.Intent != string(agentruntime.IntentRouteDraft) {
		t.Fatalf("expected route_draft intent, got %s", run.Intent)
	}
	if run.Mode != string(agentruntime.ExecutionModeWorkflow) {
		t.Fatalf("expected workflow mode, got %s", run.Mode)
	}
	if run.Model != "qwen-plus" {
		t.Fatalf("expected model qwen-plus, got %s", run.Model)
	}
	if run.Status != string(agentruntime.RunStatusRunning) {
		t.Fatalf("expected running status, got %s", run.Status)
	}
	if run.InputSummary != "杭州两日游，预算1500，喜欢美食" {
		t.Fatalf("unexpected input summary: %s", run.InputSummary)
	}

	var _ *memory.AgentRun = run
}
