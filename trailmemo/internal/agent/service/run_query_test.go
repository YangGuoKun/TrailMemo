package service

import (
	"testing"
	"time"

	"github.com/trailmemo/internal/agent/memory"
)

func TestBuildRunDetailResponse(t *testing.T) {
	now := time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC)
	entityID := uint64(99)
	run := &memory.AgentRun{
		RunID:       "run-1",
		UserID:      7,
		SessionID:   "session-1",
		Intent:      "route_draft",
		Mode:        "workflow",
		Status:      "completed",
		Model:       "qwen-plus",
		TotalTokens: 1234,
		LatencyMs:   5678,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	steps := []memory.AgentStep{{
		StepIdx:   1,
		StepType:  "validation",
		Name:      "input_guardrail",
		Status:    "success",
		LatencyMs: 12,
		CreatedAt: now,
	}}
	artifacts := []memory.AgentArtifact{{
		ArtifactID:          "artifact-1",
		Type:                "route_draft",
		Status:              "committed",
		CommittedEntityType: "create_route",
		CommittedEntityID:   &entityID,
		CreatedAt:           now,
	}}

	resp := buildRunDetailResponse(run, steps, artifacts)

	if resp.RunID != "run-1" || resp.Intent != "route_draft" || resp.Status != "completed" {
		t.Fatalf("unexpected run fields: %+v", resp)
	}
	if resp.TotalTokens != 1234 || resp.LatencyMs != 5678 {
		t.Fatalf("unexpected metrics: %+v", resp)
	}
	if len(resp.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(resp.Steps))
	}
	if resp.Steps[0].Index != 1 || resp.Steps[0].Name != "input_guardrail" {
		t.Fatalf("unexpected step: %+v", resp.Steps[0])
	}
	if len(resp.Artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(resp.Artifacts))
	}
	if resp.Artifacts[0].ArtifactID != "artifact-1" || resp.Artifacts[0].CommittedEntityID == nil || *resp.Artifacts[0].CommittedEntityID != 99 {
		t.Fatalf("unexpected artifact: %+v", resp.Artifacts[0])
	}
}
