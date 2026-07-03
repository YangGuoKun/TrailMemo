package service

import (
	"context"

	"github.com/trailmemo/internal/agent/memory"
	"github.com/trailmemo/internal/agent/observability"
)

// GetPreferences 获取用户 AI 画像快照。
func (s *AgentService) GetPreferences(ctx context.Context, userID uint64) *memory.PreferenceSnapshot {
	return s.prefStore.GetSnapshot(ctx, userID)
}

// UpdatePreferences 用户手动更新 AI 偏好（显式优先）。
func (s *AgentService) UpdatePreferences(ctx context.Context, userID uint64, update *memory.PreferenceUpdate) error {
	return s.prefStore.UpdateExplicit(ctx, userID, update)
}

// DeleteMemory 清空用户 AI 记忆。
func (s *AgentService) DeleteMemory(ctx context.Context, userID uint64) error {
	return s.prefStore.DeleteMemory(ctx, userID)
}

// GetMetrics 返回 Agent 运行指标快照。
func (s *AgentService) GetMetrics() interface{} {
	return observability.GetMetrics().Snapshot()
}

// RecordBehaviorSignal 记录行为信号（供其他 service 在业务操作后调用）。
func (s *AgentService) RecordBehaviorSignal(ctx context.Context, userID uint64, signal memory.BehaviorSignal) {
	_ = s.prefStore.RecordSignal(ctx, userID, signal)
}
