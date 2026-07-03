package service

import (
	"context"
	"fmt"

	"github.com/trailmemo/internal/agent/dto"
	"github.com/trailmemo/internal/agent/memory"
	agentruntime "github.com/trailmemo/internal/agent/runtime"
	"github.com/trailmemo/internal/agent/workflow"
)

// CreateRouteDraft 执行路线草稿生成 Workflow。
// 对应设计文档 §7 —— 一句话 → 结构化路线草稿。
func (s *AgentService) CreateRouteDraft(ctx context.Context, userID uint64, req *dto.RouteDraftRequest) (*dto.RouteDraftResponse, error) {
	return s.CreateRouteDraftWithSession(ctx, userID, &dto.RouteDraftRequest{
		Query: req.Query, StartCity: req.StartCity, TargetCity: req.TargetCity,
		Days: req.Days, Budget: req.Budget, TravelStyles: req.TravelStyles,
	})
}

// CreateRouteDraftWithSession 执行路线草稿生成并保存会话记录。
func (s *AgentService) CreateRouteDraftWithSession(ctx context.Context, userID uint64, req *dto.RouteDraftRequest) (*dto.RouteDraftResponse, error) {
	if !s.cfg.Enabled {
		return nil, fmt.Errorf("agent 未启用")
	}

	wc := s.buildWorkflowContext(ctx, userID, agentruntime.IntentRouteDraft, agentruntime.ExecutionModeWorkflow)
	wc.Input = &workflow.RouteDraftRequest{
		Query: req.Query, StartCity: req.StartCity, TargetCity: req.TargetCity,
		Days: req.Days, Budget: req.Budget, TravelStyles: req.TravelStyles,
	}
	if err := s.startWorkflowRun(ctx, wc); err != nil {
		return nil, fmt.Errorf("创建运行记录失败: %w", err)
	}

	wf := &workflow.RouteDraftWorkflow{}
	result, err := wf.Run(wc)
	if err != nil {
		s.failWorkflowRun(ctx, wc, "AGENT_WORKFLOW_FAILED", err)
		return nil, err
	}
	s.completeWorkflowRun(ctx, wc, result)

	artifact := result.Artifact.(*workflow.RouteDraftArtifact)
	resp := &dto.RouteDraftResponse{
		RunID: wc.RunID, ArtifactID: result.ArtifactID,
		RouteDraft: convertRouteDraftToDTO(artifact),
		Warnings:   result.Warnings, ApprovalRequired: result.Approval, NextAction: result.NextAction,
	}

	if req.SessionID != "" && s.sessMem != nil {
		assistantContent := fmt.Sprintf("已为你生成路线「%s」（%s→%s），共%d个打卡点，预估费用%.0f元。",
			resp.RouteDraft.Title, resp.RouteDraft.StartCity, resp.RouteDraft.EndCity,
			len(resp.RouteDraft.Checkpoints), resp.RouteDraft.EstimatedBudget)
		_ = s.sessMem.AppendMessages(ctx, req.SessionID, []memory.SessionMessage{
			{Role: "user", Content: req.Query},
			{Role: "assistant", Content: assistantContent},
		}, 50)
		s.EnsureSession(ctx, userID, req.SessionID, req.Query, "")
	}

	return resp, nil
}
