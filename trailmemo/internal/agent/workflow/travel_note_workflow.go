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

// TravelNoteRequest 是游记生成请求。
type TravelNoteRequest struct {
	RouteID              uint64 `json:"route_id" binding:"required"`  // 路线 ID
	Style                string `json:"style"`                        // 风格：story/journal/guide/social/poetic
	IncludeCheckinContent bool   `json:"include_checkin_content"`     // 是否包含打卡感受
	IncludeImages        bool   `json:"include_images"`              // 是否引用打卡图片
}

// TravelNoteArtifact 是游记草稿产物。
type TravelNoteArtifact struct {
	Type              string   `json:"type"`                // "travel_note_draft"
	RouteID           uint64   `json:"route_id"`            // 来源路线 ID
	Title             string   `json:"title"`               // 标题（8-32字）
	Content           string   `json:"content"`             // 正文
	Style             string   `json:"style"`               // 风格
	SuggestedTags     []string `json:"suggested_tags"`      // 推荐标签（3-8个）
	ImageRefs         []string `json:"image_refs,omitempty"` // 引用的打卡图片
	SourceCheckinIDs  []uint64 `json:"source_checkin_ids"`  // 来源打卡 ID
}

// TravelNoteWorkflow 实现"从路线+打卡生成游记草稿"的固定步骤编排。
// 对应 Phase 3 设计文档 §9。
type TravelNoteWorkflow struct{}

func (w *TravelNoteWorkflow) Name() string { return "travel_note" }

func (w *TravelNoteWorkflow) Run(wc *WorkflowContext) (*WorkflowResult, error) {
	req := wc.Input.(*TravelNoteRequest)
	start := time.Now()
	warnings := make([]string, 0)
	totalTokens := 0

	// Step 1: 输入校验
	if err := wc.Guardrail.CheckInput(wc.Ctx, fmt.Sprintf("生成路线%d的游记", req.RouteID)); err != nil {
		return nil, fmt.Errorf("输入校验失败: %w", err)
	}
	wc.addStep(wc.RunID, 1, "validation", "input_guardrail", "success", 0)

	// Step 2: 调工具获取路线详情
	routeResult, err := wc.ToolRegistry.Execute(wc.Ctx, "checkin.list_by_route",
		json.RawMessage(fmt.Sprintf(`{"route_id":%d,"page":1,"size":50}`, req.RouteID)))
	if err != nil || !routeResult.Success {
		return nil, fmt.Errorf("获取打卡记录失败，无法生成游记")
	}
	wc.addStep(wc.RunID, 2, "tool", "checkin.list_by_route", "success", 0)

	// 从工具结果提取打卡数据
	checkinsData, _ := json.Marshal(routeResult.Data)
	var checkinSummaries []string
	imageRefs := make([]string, 0)
	sourceIDs := make([]uint64, 0)

	// 解析打卡 JSON（工具返回的是通用 interface{}，这里从原始 JSON 提取）
	var rawData struct {
		Checkins []struct {
			ID       uint64 `json:"id"`
			Content  string `json:"content"`
			PhotoURL string `json:"photo_url"`
			Rating   int    `json:"rating"`
		} `json:"checkins"`
	}
	json.Unmarshal(checkinsData, &rawData)
	for _, c := range rawData.Checkins {
		if c.Content != "" {
			checkinSummaries = append(checkinSummaries, fmt.Sprintf("打卡：%s（评分%d）", c.Content, c.Rating))
		}
		if c.PhotoURL != "" && req.IncludeImages {
			imageRefs = append(imageRefs, c.PhotoURL)
		}
		sourceIDs = append(sourceIDs, c.ID)
	}

	if len(checkinSummaries) == 0 {
		return nil, fmt.Errorf("该路线暂无打卡内容，无法生成游记")
	}

	// Step 3: LLM 生成游记
	style := req.Style
	if style == "" { style = "story" }
	draftPrompt := fmt.Sprintf(
		`你是旅行游记作家。根据以下打卡记录，写一篇%s风格的游记。
打卡记录：%v
生成JSON：{"title":"标题(8-32字)","content":"正文内容","suggested_tags":["标签1","标签2"]}
标签3-8个，不包含敏感词。内容长度适合社区分享。只输出JSON。`, style, checkinSummaries)

	draftResp, err := llm.Retry(wc.Ctx, llm.DefaultRetryConfig(), "generate_travel_note",
		func(ctx context.Context) (string, error) {
			r, e := wc.LLMClient.Chat(ctx, llm.ChatRequest{
				Messages: []llm.Message{{Role: "user", Content: draftPrompt}},
				MaxTokens: 2000,
			})
			if e != nil { return "", e }
			return r.Content, nil
		})
	if err != nil {
		return nil, fmt.Errorf("游记生成失败: %w", err)
	}
	totalTokens += 2000
	wc.addStep(wc.RunID, 3, "llm", "generate_artifact", "success", 0)

	// Step 4+5: 输出校验 + 保存
	var artifact TravelNoteArtifact
	if err := json.Unmarshal([]byte(draftResp), &artifact); err != nil {
		return nil, fmt.Errorf("解析游记草稿失败: %w", err)
	}
	artifact.Type = "travel_note_draft"
	artifact.RouteID = req.RouteID
	artifact.Style = style
	artifact.ImageRefs = imageRefs
	artifact.SourceCheckinIDs = sourceIDs

	artID := uuid.NewString()
	contentJSON, _ := json.Marshal(artifact)
	if err := wc.ArtifactStore.SaveArtifact(wc.Ctx, &memory.AgentArtifact{
		ArtifactID: artID, RunID: wc.RunID, UserID: wc.UserID,
		Type: "travel_note_draft", ContentJSON: string(contentJSON),
	}); err != nil {
		return nil, fmt.Errorf("保存产物失败: %w", err)
	}
	wc.addStep(wc.RunID, 4, "artifact", "save_artifact", "success", 0)

	return &WorkflowResult{
		Artifact: &artifact, ArtifactType: "travel_note_draft", ArtifactID: artID,
		TotalTokens: totalTokens, LatencyMs: time.Since(start).Milliseconds(),
		Warnings: warnings, NextAction: "commit_to_post", Approval: true,
	}, nil
}
