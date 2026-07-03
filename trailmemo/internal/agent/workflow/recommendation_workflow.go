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
)

// RecommendRequest 是推荐请求。
type RecommendRequest struct {
	Query      string   `json:"query" binding:"required"` // 用户需求描述
	Days       int      `json:"days"`                     // 出行天数
	Budget     int      `json:"budget"`                   // 预算
	Interests  []string `json:"interests"`                // 兴趣标签
	TravelType string   `json:"travel_type"`              // 旅行类型
}

// RecommendItem 是一条推荐条目。
type RecommendItem struct {
	Title           string   `json:"title"`                      // 推荐标题
	City            string   `json:"city"`                       // 推荐城市
	Reason          string   `json:"reason"`                     // 推荐理由
	EstimatedBudget int      `json:"estimated_budget"`           // 估算费用
	Days            int      `json:"days"`                       // 建议天数
	Tags            []string `json:"tags"`                       // 标签
	SourceRouteIDs  []uint64 `json:"source_route_ids,omitempty"` // 参考路线 ID
}

// RecommendArtifact 是推荐结果产物。
type RecommendArtifact struct {
	Type  string          `json:"type"`  // "recommendation"
	Query string          `json:"query"` // 原始查询
	Items []RecommendItem `json:"items"` // 推荐列表
}

// RecommendWorkflow 实现智能推荐编排。
// 对应 Phase 3 设计文档 §6。
type RecommendWorkflow struct{}

func (w *RecommendWorkflow) Name() string { return "recommend" }

func (w *RecommendWorkflow) Run(wc *WorkflowContext) (*WorkflowResult, error) {
	req := wc.Input.(*RecommendRequest)
	start := time.Now()
	warnings := make([]string, 0)
	totalTokens := 0

	// Step 1: 输入校验
	if err := wc.Guardrail.CheckInput(wc.Ctx, req.Query); err != nil {
		return nil, fmt.Errorf("输入校验失败: %w", err)
	}
	wc.addStep(wc.RunID, 1, "validation", "input_guardrail", "success", 0)

	// Step 2: 搜索候选路线——根据用户需求过滤，结果作为灵感参考
	searchCity := ""
	searchKeyword := ""

	if len(req.Interests) > 0 {
		searchKeyword = strings.Join(req.Interests, " ")
	}

	if req.TravelType != "" {
		if searchKeyword != "" {
			searchKeyword += " " + req.TravelType
		} else {
			searchKeyword = req.TravelType
		}
	}

	searchArgs := map[string]interface{}{"limit": 5}
	if searchCity != "" {
		searchArgs["city"] = searchCity
	}
	if searchKeyword != "" {
		searchArgs["keyword"] = searchKeyword
	}
	searchArgsJSON, _ := json.Marshal(searchArgs)

	routeResult, routeErr := wc.ToolRegistry.Execute(wc.Ctx, "route.search_public",
		json.RawMessage(searchArgsJSON))
	referenceRoutes := ""
	if routeErr != nil {
		warnings = append(warnings, "公开路线搜索失败，推荐结果无路线参考")
	} else {
		referenceRoutes = extractToolContext(routeResult)
	}
	wc.addStep(wc.RunID, 2, "tool", "route.search_public", "success", 0)

	// Step 3: LLM 生成推荐——基于用户需求生成原创推荐，已有路线仅作灵感
	prefSummary := "无"
	if wc.Preferences != nil && !wc.Preferences.IsEmpty() {
		prefSummary = fmt.Sprintf("预算：%s，节奏：%s，兴趣：%v", wc.Preferences.BudgetLevel, wc.Preferences.Pace, wc.Preferences.Interests)
	}
	recPrompt := fmt.Sprintf(
		`你是旅行推荐顾问。请根据用户需求生成 3 条原创旅行推荐。
用户需求：%s，天数%d，预算%d，兴趣：%v，旅行类型：%s。
用户历史偏好：%s

平台已有路线（仅供灵感参考，不要照搬）：%s

要求：
1. 每条推荐必须是原创的，不要直接使用平台已有路线的内容
2. 推荐理由必须与用户需求和偏好强相关
3. 推荐城市应与用户需求匹配，如避暑推荐凉爽城市
4. 预算和天数估算要合理
5. 标签应准确反映推荐特点

输出JSON数组：[{"title":"原创推荐标题","city":"推荐城市","reason":"详细推荐理由，说明为什么适合用户","estimated_budget":0,"days":0,"tags":["标签1","标签2"]}]
只输出JSON数组。`, req.Query, req.Days, req.Budget, req.Interests, req.TravelType, prefSummary, referenceRoutes)

	recResp, err := llm.Retry(wc.Ctx, llm.DefaultRetryConfig(), "generate_recommendation",
		func(ctx context.Context) (string, error) {
			r, e := wc.LLMClient.Chat(ctx, llm.ChatRequest{
				Messages:    []llm.Message{{Role: "user", Content: recPrompt}},
				MaxTokens:   1500,
				Temperature: 0.9,
			})
			if e != nil {
				return "", e
			}
			return r.Content, nil
		})
	if err != nil {
		return nil, fmt.Errorf("推荐生成失败: %w", err)
	}
	totalTokens += 1500
	wc.addStep(wc.RunID, 3, "llm", "generate_artifact", "success", 0)

	// 解析 + 保存
	var items []RecommendItem
	if err := json.Unmarshal([]byte(recResp), &items); err != nil {
		return nil, fmt.Errorf("解析推荐结果失败: %w", err)
	}

	artifact := RecommendArtifact{Type: "recommendation", Query: req.Query, Items: items}
	artID := uuid.NewString()
	contentJSON, _ := json.Marshal(artifact)
	_ = wc.ArtifactStore.SaveArtifact(wc.Ctx, &memory.AgentArtifact{
		ArtifactID: artID, RunID: wc.RunID, UserID: wc.UserID,
		Type: "recommendation", ContentJSON: string(contentJSON),
	})
	wc.addStep(wc.RunID, 4, "artifact", "save_artifact", "success", 0)

	return &WorkflowResult{
		Artifact: &artifact, ArtifactType: "recommendation", ArtifactID: artID,
		TotalTokens: totalTokens, LatencyMs: time.Since(start).Milliseconds(),
		Warnings: warnings, NextAction: "view_detail", Approval: false,
	}, nil
}
