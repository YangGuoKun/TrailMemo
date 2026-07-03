package service

import (
	"context"
	"fmt"

	agentruntime "github.com/trailmemo/internal/agent/runtime"
	"github.com/trailmemo/internal/agent/dto"
	"github.com/trailmemo/internal/agent/workflow"
)

// RemixRoute 基于公开路线改造生成新路线草稿。
// 对应设计文档 §10 —— 读取公开路线 → 提取改造约束 → 生成改造草稿。
func (s *AgentService) RemixRoute(ctx context.Context, userID uint64, routeIDStr string, req *dto.RemixRequest) (*dto.RemixResponse, error) {
	if !s.cfg.Enabled { return nil, fmt.Errorf("agent 未启用") }

	wc := s.buildWorkflowContext(ctx, userID, agentruntime.IntentRouteRemix, agentruntime.ExecutionModeWorkflow)
	wc.Input = &workflow.RemixRequest{
		SourceRouteID: parseUint64(routeIDStr), Query: req.Query,
		Days: req.Days, Budget: req.Budget, TravelStyles: req.TravelStyles,
	}
	if err := s.startWorkflowRun(ctx, wc); err != nil {
		return nil, fmt.Errorf("创建运行记录失败: %w", err)
	}

	wf := &workflow.RouteRemixWorkflow{}
	result, err := wf.Run(wc)
	if err != nil {
		s.failWorkflowRun(ctx, wc, "AGENT_WORKFLOW_FAILED", err)
		return nil, err
	}
	s.completeWorkflowRun(ctx, wc, result)

	remix := result.Artifact.(*workflow.RemixArtifact)
	changes := make([]dto.RemixChangeItem, 0, len(remix.ChangeSummary))
	for _, c := range remix.ChangeSummary {
		changes = append(changes, dto.RemixChangeItem{Action: c.Action, Point: c.Point, Reason: c.Reason})
	}
	return &dto.RemixResponse{
		RunID: wc.RunID, SourceRouteID: remix.SourceRouteID, ArtifactID: result.ArtifactID,
		ChangeSummary: changes, RouteDraft: convertRemixToDTO(remix),
		Warnings: result.Warnings, ApprovalRequired: result.Approval, NextAction: result.NextAction,
	}, nil
}
