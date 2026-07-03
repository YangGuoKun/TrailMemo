package workflow

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/trailmemo/internal/agent/guardrail"
	"github.com/trailmemo/internal/agent/llm"
	"github.com/trailmemo/internal/agent/memory"
	"github.com/trailmemo/internal/agent/runtime"
	"github.com/trailmemo/internal/agent/tools"
	"github.com/trailmemo/internal/config"
)

type fakeRouteDraftLLM struct {
	calls []llm.ChatRequest
}

func (f *fakeRouteDraftLLM) Chat(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	f.calls = append(f.calls, req)
	if len(f.calls) == 1 {
		return &llm.ChatResponse{
			Content: `{"start_city":"上海","end_city":"杭州","days":2,"budget":1500,"style":"轻松","interests":["美食","文化"]}`,
			Usage:   llm.Usage{PromptTokens: 30, CompletionTokens: 20, TotalTokens: 50},
		}, nil
	}
	if !strings.Contains(req.Messages[0].Content, "候选POI打卡点") {
		return nil, nil
	}
	return &llm.ChatResponse{
		Content: `{"title":"杭州两日逛吃线","summary":"轻松低强度","start_city":"上海","end_city":"杭州","estimated_budget":1400,"estimated_hours":16,"checkpoints":[{"name":"河坊街","city":"杭州","address":"河坊街","sequence":1,"arrive_time":"Day1 10:00","stay_duration":90,"description":"逛吃"}]}`,
		Usage:   llm.Usage{PromptTokens: 300, CompletionTokens: 120, TotalTokens: 420},
	}, nil
}

type fakeWorkflowTool struct {
	name    string
	data    map[string]interface{}
	success bool
}

func (t fakeWorkflowTool) Name() string        { return t.name }
func (t fakeWorkflowTool) Description() string { return t.name }
func (t fakeWorkflowTool) Permission() tools.Permission {
	return tools.PermissionRead
}
func (t fakeWorkflowTool) JSONSchema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"},"keyword":{"type":"string"},"limit":{"type":"integer"}}}`)
}
func (t fakeWorkflowTool) Execute(ctx context.Context, args json.RawMessage) (*tools.ToolResult, error) {
	if !t.success {
		return &tools.ToolResult{Success: false, Error: "tool failed"}, nil
	}
	return &tools.ToolResult{Success: true, Data: t.data}, nil
}

type fakeRunRecorder struct {
	steps []memory.AgentStep
}

func (r *fakeRunRecorder) CreateRun(ctx context.Context, run *memory.AgentRun) error {
	return nil
}

func (r *fakeRunRecorder) CompleteRun(ctx context.Context, runID string, totalTokens int, latencyMs int64) error {
	return nil
}

func (r *fakeRunRecorder) FailRun(ctx context.Context, runID string, errorCode string, errorMsg string) error {
	return nil
}

func (r *fakeRunRecorder) AddStep(ctx context.Context, step *memory.AgentStep) error {
	r.steps = append(r.steps, *step)
	return nil
}

type fakeArtifactRecorder struct {
	artifacts []memory.AgentArtifact
}

func (r *fakeArtifactRecorder) SaveArtifact(ctx context.Context, artifact *memory.AgentArtifact) error {
	r.artifacts = append(r.artifacts, *artifact)
	return nil
}

func TestRouteDraftWorkflowRunsFullDraftFlow(t *testing.T) {
	reg := tools.NewRegistry()
	reg.Register(fakeWorkflowTool{name: "route.search_public", success: true, data: map[string]interface{}{
		"routes": []map[string]interface{}{{"title": "西湖慢游"}},
	}})
	reg.Register(fakeWorkflowTool{name: "map.poi_search", success: true, data: map[string]interface{}{
		"pois": []map[string]interface{}{{"name": "河坊街", "city": "杭州"}},
	}})
	llmClient := &fakeRouteDraftLLM{}
	runRecorder := &fakeRunRecorder{}
	artifactRecorder := &fakeArtifactRecorder{}

	wc := &WorkflowContext{
		Ctx:           context.Background(),
		RunID:         "run_test",
		UserID:        7,
		Intent:        runtime.IntentRouteDraft,
		Mode:          runtime.ExecutionModeWorkflow,
		Input:         &RouteDraftRequest{Query: "周末去杭州两天，想吃好吃的", Days: 2, Budget: 1500},
		LLMClient:     llmClient,
		ToolRegistry:  reg,
		RunStore:      runRecorder,
		ArtifactStore: artifactRecorder,
		Guardrail:     guardrail.NewService(testAgentConfig()),
	}

	result, err := (&RouteDraftWorkflow{}).Run(wc)
	if err != nil {
		t.Fatalf("expected workflow success, got %v", err)
	}
	if result.ArtifactType != "route_draft" {
		t.Fatalf("unexpected artifact type: %s", result.ArtifactType)
	}
	if result.TotalTokens != 470 {
		t.Fatalf("expected real token usage 470, got %d", result.TotalTokens)
	}
	if len(artifactRecorder.artifacts) != 1 {
		t.Fatalf("expected one saved artifact, got %d", len(artifactRecorder.artifacts))
	}
	if len(runRecorder.steps) != 7 {
		t.Fatalf("expected 7 recorded steps, got %d", len(runRecorder.steps))
	}
	for _, step := range runRecorder.steps {
		if step.Status != "success" {
			t.Fatalf("expected all steps success, got %+v", step)
		}
	}
	if len(llmClient.calls) != 2 {
		t.Fatalf("expected two llm calls, got %d", len(llmClient.calls))
	}
	if !strings.Contains(llmClient.calls[1].Messages[0].Content, "河坊街") {
		t.Fatalf("expected POI context in draft prompt: %s", llmClient.calls[1].Messages[0].Content)
	}
}

func TestRouteDraftWorkflowMarksFailedToolStep(t *testing.T) {
	reg := tools.NewRegistry()
	reg.Register(fakeWorkflowTool{name: "route.search_public", success: false})
	reg.Register(fakeWorkflowTool{name: "map.poi_search", success: true, data: map[string]interface{}{
		"pois": []map[string]interface{}{{"name": "河坊街", "city": "杭州"}},
	}})
	runRecorder := &fakeRunRecorder{}

	wc := &WorkflowContext{
		Ctx:           context.Background(),
		RunID:         "run_failed_tool",
		UserID:        7,
		Intent:        runtime.IntentRouteDraft,
		Mode:          runtime.ExecutionModeWorkflow,
		Input:         &RouteDraftRequest{Query: "周末去杭州两天，想吃好吃的", Days: 2, Budget: 1500},
		LLMClient:     &fakeRouteDraftLLM{},
		ToolRegistry:  reg,
		RunStore:      runRecorder,
		ArtifactStore: &fakeArtifactRecorder{},
		Guardrail:     guardrail.NewService(testAgentConfig()),
	}

	_, err := (&RouteDraftWorkflow{}).Run(wc)
	if err != nil {
		t.Fatalf("expected workflow to degrade and continue, got %v", err)
	}
	if len(runRecorder.steps) < 3 {
		t.Fatalf("expected route search step, got %+v", runRecorder.steps)
	}
	routeStep := runRecorder.steps[2]
	if routeStep.Name != "route.search_public" || routeStep.Status != "failed" {
		t.Fatalf("expected failed route tool step, got %+v", routeStep)
	}
}

func testAgentConfig() config.AgentConfig {
	return config.AgentConfig{MaxInputLength: 5000}
}
