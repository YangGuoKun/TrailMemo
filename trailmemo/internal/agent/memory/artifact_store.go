package memory

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/trailmemo/internal/config"
	"github.com/trailmemo/internal/platform/logger"
	"go.uber.org/zap"
)

// ArtifactStore 管理 Agent 产物的 MySQL 持久化。
// 产物包括路线草稿、游记草稿、推荐结果等，支持 draft→committed 状态流转。
type ArtifactStore struct{}

func NewArtifactStore() *ArtifactStore { return &ArtifactStore{} }

// SaveArtifact 保存一个新的 Agent 产物，自动生成 artifact_id。
func (s *ArtifactStore) SaveArtifact(ctx context.Context, artifact *AgentArtifact) error {
	if artifact.ArtifactID == "" {
		artifact.ArtifactID = uuid.NewString()
	}
	artifact.Status = "draft"
	db := config.GetDB().WithContext(ctx)
	if err := db.Create(artifact).Error; err != nil {
		logger.FromContext(ctx).Error("artifact_save_failed", zap.Error(err))
		return fmt.Errorf("保存产物失败: %w", err)
	}
	logger.FromContext(ctx).Info("artifact_saved",
		zap.String("artifact_id", artifact.ArtifactID),
		zap.String("type", artifact.Type))
	return nil
}

// CommitArtifact 将草稿状态改为 committed，关联业务实体 ID。
// 对应设计文档 ADR-004：只有用户确认后才提交。
func (s *ArtifactStore) CommitArtifact(ctx context.Context, artifactID string, entityType string, entityID uint64) error {
	db := config.GetDB().WithContext(ctx)
	return db.Model(&AgentArtifact{}).Where("artifact_id = ? AND status IN ?", artifactID, []string{"draft", "approved"}).Updates(map[string]interface{}{
		"status":                "committed",
		"committed_entity_type": entityType,
		"committed_entity_id":   entityID,
	}).Error
}

// ApproveArtifact 将待确认产物标记为 approved，供后续 commit 使用。
func (s *ArtifactStore) ApproveArtifact(ctx context.Context, artifactID string, userID uint64) error {
	db := config.GetDB().WithContext(ctx)
	return db.Model(&AgentArtifact{}).
		Where("artifact_id = ? AND user_id = ? AND status = ?", artifactID, userID, "draft").
		Update("status", "approved").Error
}

// GetArtifactByID 根据 artifact_id 查询产物。
func (s *ArtifactStore) GetArtifactByID(ctx context.Context, artifactID string) (*AgentArtifact, error) {
	var a AgentArtifact
	db := config.GetDB().WithContext(ctx)
	if err := db.Where("artifact_id = ?", artifactID).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

// GetArtifactsByRunID 查询某次运行的全部产物。
func (s *ArtifactStore) GetArtifactsByRunID(ctx context.Context, runID string) ([]AgentArtifact, error) {
	var artifacts []AgentArtifact
	db := config.GetDB().WithContext(ctx)
	if err := db.Where("run_id = ?", runID).Order("created_at DESC").Find(&artifacts).Error; err != nil {
		return nil, err
	}
	return artifacts, nil
}

// IsCommitted 检查产物是否已被提交（幂等性保证，防止重复创建路线或帖子）。
func (s *ArtifactStore) IsCommitted(ctx context.Context, artifactID string) (bool, error) {
	var a AgentArtifact
	db := config.GetDB().WithContext(ctx)
	if err := db.Where("artifact_id = ?", artifactID).First(&a).Error; err != nil {
		return false, err
	}
	return a.Status == "committed", nil
}
