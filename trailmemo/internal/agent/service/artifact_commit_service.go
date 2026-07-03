package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/trailmemo/internal/agent/dto"
	"github.com/trailmemo/internal/agent/workflow"
	"github.com/trailmemo/internal/model"
	"github.com/trailmemo/internal/platform/logger"
	agentservice2 "github.com/trailmemo/internal/service"
	"go.uber.org/zap"
)

// CommitArtifact 执行产物提交：校验审批→幂等检查→调用业务 Service 创建真实实体。
// 对应设计文档 §8 —— 用户确认后将草稿导入为真实 Route 或 Post。
func (s *AgentService) CommitArtifact(ctx context.Context, userID uint64, artifactID string, req *dto.ArtifactCommitRequest) (*dto.ArtifactCommitResponse, error) {
	if !s.cfg.Enabled { return nil, fmt.Errorf("agent 未启用") }

	// 幂等检查
	if existing, _ := s.guardrail.CheckIdempotency(ctx, userID, req.IdempotencyKey); existing != "" {
		return nil, fmt.Errorf("重复请求：idempotency_key 已使用")
	}

	// 加载 artifact
	art, err := s.artStore.GetArtifactByID(ctx, artifactID)
	if err != nil { return nil, fmt.Errorf("产物不存在: %s", artifactID) }
	if art.Status == "committed" {
		return &dto.ArtifactCommitResponse{ArtifactID: artifactID, Status: "committed"}, nil
	}
	needsApproval := s.guardrail.NeedsApproval(req.CommitType)
	if !canCommitArtifactStatus(art.Status, needsApproval) {
		return nil, fmt.Errorf("产物状态不允许提交: %s", art.Status)
	}

	// 审批日志
	if needsApproval {
		logger.FromContext(ctx).Info("commit_requires_approval",
			zap.String("artifact_id", artifactID), zap.String("commit_type", req.CommitType))
	}

	// 执行提交——调用业务 Service
	entityID, err := s.executeCommit(ctx, userID, req.CommitType, art.ContentJSON, req.IsPublic)
	if err != nil { return nil, err }

	if err := s.artStore.CommitArtifact(ctx, artifactID, req.CommitType, entityID); err != nil {
		return nil, fmt.Errorf("提交产物失败: %w", err) }

	_ = s.guardrail.SaveIdempotency(ctx, userID, req.IdempotencyKey,
		fmt.Sprintf(`{"artifact_id":"%s","entity_id":%d}`, artifactID, entityID))

	logger.FromContext(ctx).Info("artifact_committed",
		zap.String("artifact_id", artifactID), zap.String("commit_type", req.CommitType), zap.Uint64("entity_id", entityID))

	return &dto.ArtifactCommitResponse{ArtifactID: artifactID, Status: "committed", EntityType: req.CommitType, EntityID: entityID}, nil
}

func (s *AgentService) ApproveArtifact(ctx context.Context, userID uint64, artifactID string) (*dto.ArtifactApprovalResponse, error) {
	if !s.cfg.Enabled { return nil, fmt.Errorf("agent 未启用") }
	art, err := s.artStore.GetArtifactByID(ctx, artifactID)
	if err != nil { return nil, fmt.Errorf("产物不存在: %s", artifactID) }
	if art.UserID != userID { return nil, fmt.Errorf("产物不存在: %s", artifactID) }
	if art.Status == "approved" || art.Status == "committed" {
		return &dto.ArtifactApprovalResponse{ArtifactID: artifactID, Status: art.Status}, nil
	}
	if art.Status != "draft" {
		return nil, fmt.Errorf("产物状态不允许确认: %s", art.Status)
	}
	if err := s.artStore.ApproveArtifact(ctx, artifactID, userID); err != nil {
		return nil, fmt.Errorf("确认产物失败: %w", err)
	}
	return &dto.ArtifactApprovalResponse{ArtifactID: artifactID, Status: "approved"}, nil
}

func canCommitArtifactStatus(status string, needsApproval bool) bool {
	if needsApproval {
		return status == "approved"
	}
	return status == "draft" || status == "approved"
}

// executeCommit 根据 commit_type 调用对应的业务 Service。
func (s *AgentService) executeCommit(ctx context.Context, userID uint64, commitType, contentJSON string, isPublic int) (uint64, error) {
	switch commitType {
	case "create_route":
		return s.commitRoute(ctx, userID, contentJSON, isPublic)
	case "create_post":
		return s.commitPost(ctx, userID, contentJSON)
	default:
		return 0, fmt.Errorf("不支持的 commit_type: %s", commitType)
	}
}

func (s *AgentService) commitRoute(ctx context.Context, userID uint64, contentJSON string, isPublic int) (uint64, error) {
	var draft workflow.RouteDraftArtifact
	if err := json.Unmarshal([]byte(contentJSON), &draft); err != nil {
		return 0, fmt.Errorf("解析路线草稿失败: %w", err)
	}
	checkpoints := make([]*model.Checkpoint, 0, len(draft.Checkpoints))
	for _, cp := range draft.Checkpoints {
		checkpoints = append(checkpoints, &model.Checkpoint{
			Name: cp.Name, Description: cp.Description, City: cp.City,
			Address: cp.Address, Latitude: cp.Latitude, Longitude: cp.Longitude,
			Sequence: cp.Sequence, ArriveTime: cp.ArriveTime, StayDuration: cp.StayDuration,
		})
	}
	route, err := agentservice2.NewRouteService().CreateRoute(ctx, userID, draft.Title, draft.Summary, "",
		draft.StartCity, draft.EndCity, 0, float64(draft.EstimatedHours), isPublic, checkpoints)
	if err != nil { return 0, fmt.Errorf("创建路线失败: %w", err) }
	return route.ID, nil
}

func (s *AgentService) commitPost(ctx context.Context, userID uint64, contentJSON string) (uint64, error) {
	var note workflow.TravelNoteArtifact
	if err := json.Unmarshal([]byte(contentJSON), &note); err != nil {
		return 0, fmt.Errorf("解析游记草稿失败: %w", err)
	}
	post, err := agentservice2.NewPostService().CreatePost(ctx, userID, note.RouteID, note.Title, note.Content, "")
	if err != nil { return 0, fmt.Errorf("发布帖子失败: %w", err) }
	return post.ID, nil
}
