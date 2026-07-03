package service

import (
	"context"
	"fmt"

	agentruntime "github.com/trailmemo/internal/agent/runtime"
	"github.com/trailmemo/internal/agent/dto"
	"github.com/trailmemo/internal/agent/workflow"
)

// GenerateTravelNote 从路线打卡记录生成游记草稿。
// 对应设计文档 §9 —— 读取打卡记录 → 构建时间线 → LLM 生成游记。
func (s *AgentService) GenerateTravelNote(ctx context.Context, userID uint64, req *dto.TravelNoteRequest) (*dto.TravelNoteResponse, error) {
	if !s.cfg.Enabled { return nil, fmt.Errorf("agent 未启用") }

	wc := s.buildWorkflowContext(ctx, userID, agentruntime.IntentTravelNote, agentruntime.ExecutionModeWorkflow)
	wc.Input = &workflow.TravelNoteRequest{
		RouteID: req.RouteID, Style: req.Style,
		IncludeCheckinContent: req.IncludeCheckinContent, IncludeImages: req.IncludeImages,
	}
	if err := s.startWorkflowRun(ctx, wc); err != nil {
		return nil, fmt.Errorf("创建运行记录失败: %w", err)
	}

	wf := &workflow.TravelNoteWorkflow{}
	result, err := wf.Run(wc)
	if err != nil {
		s.failWorkflowRun(ctx, wc, "AGENT_WORKFLOW_FAILED", err)
		return nil, err
	}
	s.completeWorkflowRun(ctx, wc, result)

	note := result.Artifact.(*workflow.TravelNoteArtifact)
	return &dto.TravelNoteResponse{
		RunID: wc.RunID, ArtifactID: result.ArtifactID,
		Title: note.Title, Content: note.Content, SuggestedTags: note.SuggestedTags,
		ImageRefs: note.ImageRefs, Warnings: result.Warnings,
		ApprovalRequired: result.Approval, NextAction: result.NextAction,
	}, nil
}
