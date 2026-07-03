package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/trailmemo/internal/agent/llm"
	"github.com/trailmemo/internal/agent/memory"
	"github.com/trailmemo/internal/agent/tools"
)

// extractToolContext 将工具执行结果序列化为 JSON 字符串，用于注入 LLM prompt 上下文。
// 工具调用失败或数据为空时返回空字符串，不阻断流程。
func extractToolContext(result *tools.ToolResult) string {
	if result == nil || !result.Success || result.Data == nil {
		return ""
	}
	b, err := json.Marshal(result.Data)
	if err != nil {
		return ""
	}
	return string(b)
}

// ── Route Draft 数据结构 ─────────────────────────────

// RouteDraftRequest 是路线草稿生成的请求参数。
type RouteDraftRequest struct {
	Query        string   `json:"query" binding:"required"` // 用户自然语言需求
	StartCity    string   `json:"start_city"`               // 出发城市
	TargetCity   string   `json:"target_city"`              // 目标城市
	Days         int      `json:"days"`                     // 出行天数
	Budget       int      `json:"budget"`                   // 预算（元）
	TravelStyles []string `json:"travel_styles"`            // 旅行风格标签
}

// RouteDraftArtifact 是路线草稿产物。
// 对应设计文档 §4.2
type RouteDraftArtifact struct {
	Type            string            `json:"type"`               // "route_draft"
	Title           string            `json:"title"`              // 路线标题
	Summary         string            `json:"summary"`            // 路线概述
	StartCity       string            `json:"start_city"`         // 出发城市
	EndCity         string            `json:"end_city"`           // 目的地城市
	EstimatedBudget float64           `json:"estimated_budget"`   // 估算费用
	EstimatedHours  float64           `json:"estimated_hours"`    // 估算时长
	Confidence      float64           `json:"confidence"`         // 置信度 0-1
	Warnings        []string          `json:"warnings,omitempty"` // 警告信息
	Checkpoints     []CheckpointDraft `json:"checkpoints"`        // 打卡点列表
}

// CheckpointDraft 是打卡点草稿。
type CheckpointDraft struct {
	Name         string  `json:"name"`          // 打卡点名称
	Description  string  `json:"description"`   // 简短说明
	City         string  `json:"city"`          // 所在城市
	Address      string  `json:"address"`       // 地址
	Latitude     float64 `json:"latitude"`      // 纬度
	Longitude    float64 `json:"longitude"`     // 经度
	Sequence     int     `json:"sequence"`      // 序号，从1递增
	ArriveTime   string  `json:"arrive_time"`   // 建议到达时间
	StayDuration int     `json:"stay_duration"` // 停留分钟
}

// ── Workflow 实现 ────────────────────────────────────

// RouteDraftWorkflow 实现"一句话生成路线草稿"的固定步骤编排。
// 对应设计文档 §7。
type RouteDraftWorkflow struct{}

func (w *RouteDraftWorkflow) Name() string { return "route_draft" }

func (w *RouteDraftWorkflow) Run(wc *WorkflowContext) (*WorkflowResult, error) {
	req := wc.Input.(*RouteDraftRequest)
	start := time.Now()
	warnings := make([]string, 0)
	totalTokens := 0

	// Step 1: 输入安全校验
	if err := wc.Guardrail.CheckInput(wc.Ctx, req.Query); err != nil {
		return nil, fmt.Errorf("输入校验失败: %w", err)
	}
	wc.addStep(wc.RunID, 1, "validation", "input_guardrail", "success", 0)

	// Step 2: LLM 提取结构化约束
	type constraintsResult struct {
		StartCity string   `json:"start_city"`
		EndCity   string   `json:"end_city"`
		Days      int      `json:"days"`
		Budget    int      `json:"budget"`
		Style     string   `json:"style"`
		Interests []string `json:"interests"`
	}
	var extractedConstraints constraintsResult
	constraintsPrompt := fmt.Sprintf(
		"分析以下旅行需求，输出JSON：{start_city:string, end_city:string, days:int, budget:int, style:string, interests:[string]}。需求：%s。天数%d，预算%d。只输出JSON。",
		req.Query, req.Days, req.Budget)
	constraintsJSON := ""
	resp, err := llm.Retry(wc.Ctx, llm.DefaultRetryConfig(), "extract_constraints",
		func(ctx context.Context) (*llm.ChatResponse, error) {
			r, e := wc.LLMClient.Chat(ctx, llm.ChatRequest{
				Messages:  []llm.Message{{Role: "user", Content: constraintsPrompt}},
				MaxTokens: 200,
			})
			if e != nil {
				return nil, e
			}
			return r, nil
		})
	if err != nil {
		warnings = append(warnings, "约束提取失败，使用默认参数")
	} else {
		constraintsJSON = resp.Content
		_ = json.Unmarshal([]byte(resp.Content), &extractedConstraints)
		totalTokens += resp.Usage.TotalTokens
	}
	wc.addStep(wc.RunID, 2, "llm", "extract_constraints", stepStatus(err), 0)

	// Step 3: 调用工具——搜索公开路线作为参考（带城市筛选）
	searchCity := extractedConstraints.EndCity
	if searchCity == "" {
		searchCity = req.TargetCity
	}
	if searchCity == "" {
		searchCity = extractedConstraints.StartCity
	}
	searchKeyword := strings.Join(extractedConstraints.Interests, " ")
	if searchKeyword == "" && extractedConstraints.Style != "" {
		searchKeyword = extractedConstraints.Style
	}

	searchArgs := map[string]interface{}{"limit": 5}
	if searchCity != "" {
		searchArgs["city"] = searchCity
	}
	if searchKeyword != "" {
		searchArgs["keyword"] = searchKeyword
	}
	searchArgsJSON, _ := json.Marshal(searchArgs)

	toolResult, toolErr := wc.ToolRegistry.Execute(wc.Ctx, "route.search_public",
		json.RawMessage(searchArgsJSON))
	referenceRoutes := ""
	if toolErr != nil {
		warnings = append(warnings, "公开路线搜索失败")
	} else {
		referenceRoutes = extractToolContext(toolResult)
	}
	wc.addStep(wc.RunID, 3, "tool", "route.search_public", toolStepStatus(toolResult, toolErr), 0)

	poiArgs := map[string]interface{}{"limit": 8}
	if searchCity != "" {
		poiArgs["city"] = searchCity
	}
	if searchKeyword != "" {
		poiArgs["keyword"] = searchKeyword
	}
	poiArgsJSON, _ := json.Marshal(poiArgs)
	poiResult, poiErr := wc.ToolRegistry.Execute(wc.Ctx, "map.poi_search", json.RawMessage(poiArgsJSON))
	referencePOIs := ""
	if poiErr != nil {
		warnings = append(warnings, "POI搜索失败，路线点位可能需要手动补充")
	} else {
		referencePOIs = extractToolContext(poiResult)
	}
	wc.addStep(wc.RunID, 4, "tool", "map.poi_search", toolStepStatus(poiResult, poiErr), 0)

	// Step 4: LLM 生成路线草稿——传入约束JSON和参考路线作为上下文
	draftPrompt := fmt.Sprintf(
		`你是旅行路线规划师。根据用户需求生成旅行路线草稿。

用户需求：%s
出发城市：%s，目标城市：%s，天数：%d，预算：%d元，风格：%v
提取的约束：%s
参考公开路线：%s
候选POI打卡点：%s

重要要求：
1. 路线必须有独特性和多样性，不要完全复制参考路线
2. 打卡点应覆盖不同类型（景点、美食、文化、购物等）
3. 每个打卡点要有明确的位置和停留时间安排
4. 路线应考虑交通便利性和时间合理性

严格输出JSON：
{"title":"路线标题","summary":"概述","start_city":"出发","end_city":"目标",
"estimated_budget":费用,"estimated_hours":总时长,
"checkpoints":[{"name":"点","city":"城","address":"址","sequence":1,"arrive_time":"Day1 09:00","stay_duration":90,"description":"说明"}...]}
只输出JSON。`,
		req.Query, req.StartCity, req.TargetCity, req.Days, req.Budget, req.TravelStyles,
		constraintsJSON, referenceRoutes, referencePOIs)

	draftResp, err := llm.Retry(wc.Ctx, llm.DefaultRetryConfig(), "generate_route_draft",
		func(ctx context.Context) (*llm.ChatResponse, error) {
			r, e := wc.LLMClient.Chat(ctx, llm.ChatRequest{
				Messages:    []llm.Message{{Role: "user", Content: draftPrompt}},
				MaxTokens:   2000,
				Temperature: 0.9,
			})
			if e != nil {
				return nil, e
			}
			return r, nil
		})
	if err != nil {
		wc.addStep(wc.RunID, 5, "llm", "generate_artifact", "failed", 0)
		return nil, fmt.Errorf("路线草稿生成失败: %w", err)
	}
	totalTokens += draftResp.Usage.TotalTokens
	wc.addStep(wc.RunID, 5, "llm", "generate_artifact", "success", 0)

	// Step 6: 输出校验
	validResult, _ := wc.Guardrail.ValidateArtifactOutput(draftResp.Content)
	if !validResult.Valid {
		warnings = append(warnings, "LLM输出格式校验失败，尝试修复")
	}
	wc.addStep(wc.RunID, 6, "validation", "output_schema", "success", 0)

	// 解析产物
	artifact, err := parseRouteDraftArtifact(draftResp.Content)
	if err != nil {
		return nil, fmt.Errorf("解析路线草稿失败: %w", err)
	}

	// Step 7: 保存 artifact
	artID := uuid.NewString()
	contentJSON, _ := json.Marshal(*artifact)
	if err := wc.ArtifactStore.SaveArtifact(wc.Ctx, &memory.AgentArtifact{
		ArtifactID:  artID,
		RunID:       wc.RunID,
		UserID:      wc.UserID,
		Type:        "route_draft",
		ContentJSON: string(contentJSON),
	}); err != nil {
		return nil, fmt.Errorf("保存产物失败: %w", err)
	}
	wc.addStep(wc.RunID, 7, "artifact", "save_artifact", "success", 0)

	return &WorkflowResult{
		Artifact:     artifact,
		ArtifactType: "route_draft",
		ArtifactID:   artID,
		TotalTokens:  totalTokens,
		LatencyMs:    time.Since(start).Milliseconds(),
		Warnings:     warnings,
		NextAction:   "commit_to_route",
		Approval:     true,
	}, nil
}

func parseRouteDraftArtifact(raw string) (*RouteDraftArtifact, error) {
	jsonText, err := extractJSONObject(raw)
	if err != nil {
		return nil, err
	}

	var artifact RouteDraftArtifact
	if err := json.Unmarshal([]byte(jsonText), &artifact); err != nil {
		return nil, err
	}
	if artifact.Title == "" {
		return nil, fmt.Errorf("缺少路线标题")
	}
	if len(artifact.Checkpoints) == 0 {
		return nil, fmt.Errorf("缺少打卡点")
	}
	artifact.Type = "route_draft"
	for i := range artifact.Checkpoints {
		artifact.Checkpoints[i].Sequence = i + 1
		if artifact.Checkpoints[i].Name == "" {
			return nil, fmt.Errorf("第%d个打卡点缺少名称", i+1)
		}
	}
	return &artifact, nil
}

func extractJSONObject(raw string) (string, error) {
	text := strings.TrimSpace(raw)
	if json.Valid([]byte(text)) {
		return text, nil
	}

	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start < 0 || end <= start {
		return "", fmt.Errorf("未找到JSON对象")
	}
	candidate := text[start : end+1]
	if !json.Valid([]byte(candidate)) {
		return "", fmt.Errorf("JSON对象格式无效")
	}
	return candidate, nil
}

func stepStatus(err error) string {
	if err != nil {
		return "failed"
	}
	return "success"
}

func toolStepStatus(result *tools.ToolResult, err error) string {
	if err != nil || result == nil || !result.Success {
		return "failed"
	}
	return "success"
}

// addStep 追加一条步骤记录到 RunStore。
func (wc *WorkflowContext) addStep(runID string, idx int, stepType, name, status string, latencyMs int64) {
	if wc.RunStore == nil {
		return
	}
	_ = wc.RunStore.AddStep(wc.Ctx, &memory.AgentStep{
		RunID:     runID,
		StepIdx:   idx,
		StepType:  stepType,
		Name:      name,
		Status:    status,
		LatencyMs: latencyMs,
	})
}
