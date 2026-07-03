package service

import (
	"context"
	"fmt"

	agentruntime "github.com/trailmemo/internal/agent/runtime"
	"github.com/trailmemo/internal/agent/dto"
	"github.com/trailmemo/internal/agent/workflow"
)

// Recommend 执行推荐 Workflow。
// 对应设计文档 §6 —— 根据用户需求生成旅行推荐列表。
func (s *AgentService) Recommend(ctx context.Context, userID uint64, req *dto.RecommendRequest) (*dto.RecommendResponse, error) {
	if !s.cfg.Enabled { return nil, fmt.Errorf("agent 未启用") }

	wc := s.buildWorkflowContext(ctx, userID, agentruntime.IntentRecommend, agentruntime.ExecutionModeWorkflow)
	wc.Input = &workflow.RecommendRequest{
		Query: req.Query, Days: req.Days, Budget: req.Budget,
		Interests: req.Interests, TravelType: req.TravelType,
	}
	if err := s.startWorkflowRun(ctx, wc); err != nil {
		return nil, fmt.Errorf("创建运行记录失败: %w", err)
	}

	wf := &workflow.RecommendWorkflow{}
	result, err := wf.Run(wc)
	if err != nil {
		s.failWorkflowRun(ctx, wc, "AGENT_WORKFLOW_FAILED", err)
		return nil, err
	}
	s.completeWorkflowRun(ctx, wc, result)

	rec := result.Artifact.(*workflow.RecommendArtifact)
	items := make([]dto.RecommendItem, 0, len(rec.Items))
	for _, it := range rec.Items {
		items = append(items, dto.RecommendItem{
			Title: it.Title, City: it.City, Reason: it.Reason,
			EstimatedBudget: it.EstimatedBudget, Days: it.Days, Tags: it.Tags})
	}
	return &dto.RecommendResponse{RunID: wc.RunID, ArtifactID: result.ArtifactID, Items: items, Warnings: result.Warnings}, nil
}
