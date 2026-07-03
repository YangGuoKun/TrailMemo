package service

import (
	"context"
	"fmt"
	"time"

	"github.com/trailmemo/internal/agent/dto"
)

// ListSessions 返回用户会话列表。
func (s *AgentService) ListSessions(ctx context.Context, userID uint64) (*dto.SessionListResponse, error) {
	sessions, err := s.sessionStore.ListByUser(ctx, userID)
	if err != nil { return nil, err }
	items := make([]dto.SessionInfo, 0, len(sessions))
	for _, ss := range sessions {
		lastAt := ""
		if ss.LastMessageAt != nil { lastAt = ss.LastMessageAt.Format(time.DateTime) }
		items = append(items, dto.SessionInfo{
			SessionID: ss.SessionID, Title: ss.Title, Model: ss.Model,
			MessageCount: ss.MessageCount, LastMessageAt: lastAt,
			CreatedAt: ss.CreatedAt.Format(time.DateTime),
		})
	}
	return &dto.SessionListResponse{Sessions: items, Total: len(items)}, nil
}

// GetSession 获取会话详情（含消息历史）。
func (s *AgentService) GetSession(ctx context.Context, userID uint64, sessionID string) (*dto.SessionDetailResponse, error) {
	ss, err := s.sessionStore.GetBySessionID(ctx, sessionID)
	if err != nil { return nil, fmt.Errorf("会话不存在") }

	// 从 Redis 加载消息
	msgs, _ := s.sessMem.GetMessages(ctx, sessionID, 100)
	expired := len(msgs) == 0
	items := make([]dto.SessionMsg, 0, len(msgs))
	for _, m := range msgs {
		items = append(items, dto.SessionMsg{Role: m.Role, Content: m.Content})
	}
	return &dto.SessionDetailResponse{
		SessionID: ss.SessionID, Title: ss.Title, Messages: items,
		MessageCount: ss.MessageCount, Expired: expired,
	}, nil
}

// DeleteSession 删除会话（MySQL + Redis）。
func (s *AgentService) DeleteSession(ctx context.Context, userID uint64, sessionID string) error {
	if err := s.sessionStore.Delete(ctx, sessionID); err != nil { return err }
	_ = s.sessMem.ClearSession(ctx, sessionID)
	return nil
}

// RenameSession 重命名会话。
func (s *AgentService) RenameSession(ctx context.Context, userID uint64, sessionID, title string) error {
	return s.sessionStore.Rename(ctx, sessionID, title)
}

// EnsureSession 对话结束后自动保存/更新会话元数据（供 ChatLoop 调用）。
func (s *AgentService) EnsureSession(ctx context.Context, userID uint64, sessionID, title, model string) {
	if sessionID == "" || userID == 0 { return }
	_ = s.sessionStore.Upsert(ctx, sessionID, userID, title, model)
}
