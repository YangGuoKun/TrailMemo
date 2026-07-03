package memory

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/trailmemo/internal/config"
	"github.com/trailmemo/internal/platform/logger"
	"go.uber.org/zap"
)

// RunStore 管理 Agent 运行的 MySQL 持久化。
// 提供创建、状态更新、步骤追加和产物关联能力。
type RunStore struct{}

func NewRunStore() *RunStore { return &RunStore{} }

// CreateRun 创建一条新的 agent run 记录，返回生成的 run_id。
func (s *RunStore) CreateRun(ctx context.Context, run *AgentRun) error {
	if run.RunID == "" {
		run.RunID = uuid.NewString()
	}
	run.Status = "created"
	db := config.GetDB().WithContext(ctx)
	if err := db.Create(run).Error; err != nil {
		logger.FromContext(ctx).Error("run_store_create_failed", zap.Error(err))
		return fmt.Errorf("创建运行记录失败: %w", err)
	}
	logger.FromContext(ctx).Info("run_created",
		zap.String("run_id", run.RunID),
		zap.String("intent", run.Intent))
	return nil
}

// CompleteRun 将 run 标记为成功完成，更新耗时和 token 统计。
func (s *RunStore) CompleteRun(ctx context.Context, runID string, totalTokens int, latencyMs int64) error {
	db := config.GetDB().WithContext(ctx)
	return db.Model(&AgentRun{}).Where("run_id = ?", runID).Updates(map[string]interface{}{
		"status":       "completed",
		"total_tokens": totalTokens,
		"latency_ms":   latencyMs,
	}).Error
}

// FailRun 将 run 标记为失败，记录错误码和脱敏错误信息。
func (s *RunStore) FailRun(ctx context.Context, runID, errorCode, errorMsg string) error {
	db := config.GetDB().WithContext(ctx)
	return db.Model(&AgentRun{}).Where("run_id = ?", runID).Updates(map[string]interface{}{
		"status":        "failed",
		"error_code":    errorCode,
		"error_message": errorMsg,
	}).Error
}

// AddStep 向指定 run 追加一条步骤记录。
func (s *RunStore) AddStep(ctx context.Context, step *AgentStep) error {
	if step.InputJSON == "" {
		step.InputJSON = "{}"
	}
	if step.OutputJSON == "" {
		step.OutputJSON = "{}"
	}
	db := config.GetDB().WithContext(ctx)
	if err := db.Create(step).Error; err != nil {
		logger.FromContext(ctx).Error("step_create_failed", zap.Error(err), zap.String("run_id", step.RunID))
		return fmt.Errorf("记录步骤失败: %w", err)
	}
	return nil
}

// GetRunByID 根据 run_id 查询运行记录。
func (s *RunStore) GetRunByID(ctx context.Context, runID string) (*AgentRun, error) {
	var run AgentRun
	db := config.GetDB().WithContext(ctx)
	if err := db.Where("run_id = ?", runID).First(&run).Error; err != nil {
		return nil, err
	}
	return &run, nil
}

// GetStepsByRunID 查询指定 run 的全部步骤，按 step_index 排序。
func (s *RunStore) GetStepsByRunID(ctx context.Context, runID string) ([]AgentStep, error) {
	var steps []AgentStep
	db := config.GetDB().WithContext(ctx)
	if err := db.Where("run_id = ?", runID).Order("step_index ASC").Find(&steps).Error; err != nil {
		return nil, err
	}
	return steps, nil
}
