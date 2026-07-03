package workflow

import (
	"encoding/json"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/trailmemo/internal/agent/llm"
	"github.com/trailmemo/internal/agent/memory"
)

// RemixRequest 是路线复用改造请求。
type RemixRequest struct {
	SourceRouteID uint64   `json:"-"`                   // 原路线 ID（从 URL 获取）
	Query         string   `json:"query" binding:"required"` // 改造描述
	Days          int      `json:"days"`                 // 目标天数
	Budget        int      `json:"budget"`               // 预算
	TravelStyles  []string `json:"travel_styles"`        // 目标风格
}

// RemixChangeSummary 是改造变更说明。
type RemixChangeSummary struct {
	Action  string `json:"action"`  // added/removed/modified
	Point   string `json:"point"`   // 打卡点名称
	Reason  string `json:"reason"`  // 变更原因
}

// RemixArtifact 是路线改造产物。
type RemixArtifact struct {
	Type          string               `json:"type"`           // "route_draft"
	SourceRouteID uint64               `json:"source_route_id"` // 原路线 ID
	Title         string               `json:"title"`          // 改造后标题
	Summary       string               `json:"summary"`        // 概述
	StartCity     string               `json:"start_city"`     // 出发城市
	EndCity       string               `json:"end_city"`       // 目的地城市
	EstimatedBudget float64            `json:"estimated_budget"`
	EstimatedHours  float64            `json:"estimated_hours"`
	ChangeSummary []RemixChangeSummary `json:"change_summary"` // 变更说明
	Checkpoints   []CheckpointDraft    `json:"checkpoints"`    // 改造后打卡点
}

// RouteRemixWorkflow 实现"基于公开路线改造"编排。
// 对应 Phase 3 设计文档 §10。
type RouteRemixWorkflow struct{}

func (w *RouteRemixWorkflow) Name() string { return "route_remix" }

func (w *RouteRemixWorkflow) Run(wc *WorkflowContext) (*WorkflowResult, error) {
	req := wc.Input.(*RemixRequest)
	start := time.Now()
	warnings := make([]string, 0)
	totalTokens := 0

	// Step 1: 输入校验
	if err := wc.Guardrail.CheckInput(wc.Ctx, req.Query); err != nil {
		return nil, fmt.Errorf("输入校验失败: %w", err)
	}
	wc.addStep(wc.RunID, 1, "validation", "input_guardrail", "success", 0)

	// Step 2: 约束提取——将结果注入后续生成 prompt
	extractPrompt := fmt.Sprintf(
		`分析路线改造需求，输出JSON：{targets:[要改的点], style, constraints}。需求：%s。只输出JSON。`, req.Query)
	constraintsJSON := ""
	extractResp, err := llm.Retry(wc.Ctx, llm.DefaultRetryConfig(), "extract_remix_constraints",
		func(ctx context.Context) (string, error) {
			r, e := wc.LLMClient.Chat(ctx, llm.ChatRequest{
				Messages: []llm.Message{{Role: "user", Content: extractPrompt}},
				MaxTokens: 300,
			})
			if e != nil { return "", e }
			return r.Content, nil
		})
	if err != nil {
		warnings = append(warnings, "约束提取失败")
	} else {
		constraintsJSON = extractResp
	}
	totalTokens += 300
	wc.addStep(wc.RunID, 2, "llm", "extract_constraints", "success", 0)

	// Step 3: 生成改造路线——传入提取的约束作为上下文
	remixPrompt := fmt.Sprintf(
		`你是旅行路线改造师。根据改造需求调整路线。
原路线ID：%d，改造需求：%s，目标天数%d，预算%d，风格%v。
提取的改造约束：%s
输出JSON：{"title":"","summary":"","start_city":"","end_city":"","estimated_budget":0,"estimated_hours":0,
"change_summary":[{"action":"added/removed/modified","point":"点名","reason":"原因"}],
"checkpoints":[{"name":"","city":"","sequence":1,"arrive_time":"Day1 09:00","stay_duration":90,"description":""}]}
只输出JSON。`, req.SourceRouteID, req.Query, req.Days, req.Budget, req.TravelStyles, constraintsJSON)

	remixResp, err := llm.Retry(wc.Ctx, llm.DefaultRetryConfig(), "generate_remixed_route",
		func(ctx context.Context) (string, error) {
			r, e := wc.LLMClient.Chat(ctx, llm.ChatRequest{
				Messages: []llm.Message{{Role: "user", Content: remixPrompt}},
				MaxTokens: 2000,
			})
			if e != nil { return "", e }
			return r.Content, nil
		})
	if err != nil {
		return nil, fmt.Errorf("路线改造失败: %w", err)
	}
	totalTokens += 2000
	wc.addStep(wc.RunID, 3, "llm", "generate_artifact", "success", 0)

	// 解析 + 保存
	var artifact RemixArtifact
	if err := json.Unmarshal([]byte(remixResp), &artifact); err != nil {
		return nil, fmt.Errorf("解析改造产物失败: %w", err)
	}
	artifact.Type = "route_draft"
	artifact.SourceRouteID = req.SourceRouteID
	for i := range artifact.Checkpoints {
		artifact.Checkpoints[i].Sequence = i + 1
	}

	artID := uuid.NewString()
	contentJSON, _ := json.Marshal(artifact)
	_ = wc.ArtifactStore.SaveArtifact(wc.Ctx, &memory.AgentArtifact{
		ArtifactID: artID, RunID: wc.RunID, UserID: wc.UserID,
		Type: "route_draft", ContentJSON: string(contentJSON),
	})
	wc.addStep(wc.RunID, 4, "artifact", "save_artifact", "success", 0)

	return &WorkflowResult{
		Artifact: &artifact, ArtifactType: "route_draft", ArtifactID: artID,
		TotalTokens: totalTokens, LatencyMs: time.Since(start).Milliseconds(),
		Warnings: warnings, NextAction: "commit_to_route", Approval: true,
	}, nil
}
