package service

import (
	"context"
	"errors"

	"github.com/trailmemo/internal/agent/dto"
	"github.com/trailmemo/internal/agent/memory"
)

var errRunNotFound = errors.New("运行记录不存在")

func (s *AgentService) GetRunDetail(ctx context.Context, userID uint64, runID string) (*dto.RunDetailResponse, error) {
	run, err := s.runStore.GetRunByID(ctx, runID)
	if err != nil {
		return nil, err
	}
	if run.UserID != userID {
		return nil, errRunNotFound
	}
	steps, err := s.runStore.GetStepsByRunID(ctx, runID)
	if err != nil {
		return nil, err
	}
	artifacts, err := s.artStore.GetArtifactsByRunID(ctx, runID)
	if err != nil {
		return nil, err
	}
	return buildRunDetailResponse(run, steps, artifacts), nil
}

func buildRunDetailResponse(run *memory.AgentRun, steps []memory.AgentStep, artifacts []memory.AgentArtifact) *dto.RunDetailResponse {
	resp := &dto.RunDetailResponse{
		RunID:       run.RunID,
		UserID:      run.UserID,
		SessionID:   run.SessionID,
		Intent:      run.Intent,
		Mode:        run.Mode,
		Status:      run.Status,
		Model:       run.Model,
		PromptVer:   run.PromptVer,
		TotalTokens: run.TotalTokens,
		LatencyMs:   run.LatencyMs,
		ErrorCode:   run.ErrorCode,
		ErrorMsg:    run.ErrorMsg,
		CreatedAt:   run.CreatedAt.Format(timeLayout),
		UpdatedAt:   run.UpdatedAt.Format(timeLayout),
		Steps:       make([]dto.RunStepInfo, 0, len(steps)),
		Artifacts:   make([]dto.RunArtifactInfo, 0, len(artifacts)),
	}
	for _, step := range steps {
		resp.Steps = append(resp.Steps, dto.RunStepInfo{
			Index:     step.StepIdx,
			Type:      step.StepType,
			Name:      step.Name,
			Status:    step.Status,
			LatencyMs: step.LatencyMs,
			CreatedAt: step.CreatedAt.Format(timeLayout),
		})
	}
	for _, artifact := range artifacts {
		resp.Artifacts = append(resp.Artifacts, dto.RunArtifactInfo{
			ArtifactID:          artifact.ArtifactID,
			Type:                artifact.Type,
			Status:              artifact.Status,
			CommittedEntityType: artifact.CommittedEntityType,
			CommittedEntityID:   artifact.CommittedEntityID,
			CreatedAt:           artifact.CreatedAt.Format(timeLayout),
		})
	}
	return resp
}

const timeLayout = "2006-01-02T15:04:05Z07:00"
