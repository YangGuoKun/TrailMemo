package memory

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/trailmemo/internal/config"
)

// SessionStore 管理会话元数据的 MySQL 持久化。
type SessionStore struct{}

func NewSessionStore() *SessionStore { return &SessionStore{} }

// Create 创建一条新会话记录。sessionID 为空时自动生成。
func (s *SessionStore) Create(ctx context.Context, userID uint64, title, model string) (*AgentSession, error) {
	if title == "" { title = "新对话" }
	session := &AgentSession{
		SessionID: uuid.NewString(),
		UserID:    userID,
		Title:     truncateTitle(title, 32),
		Model:     model,
		MessageCount: 1,
	}
	now := time.Now()
	session.LastMessageAt = &now

	db := config.GetDB().WithContext(ctx)
	if err := db.Create(session).Error; err != nil {
		return nil, err
	}
	return session, nil
}

// Upsert 如果 sessionID 不存在则创建，存在则更新计数和时间。
func (s *SessionStore) Upsert(ctx context.Context, sessionID string, userID uint64, title, model string) error {
	db := config.GetDB().WithContext(ctx)
	var existing AgentSession
	err := db.Where("session_id = ?", sessionID).First(&existing).Error
	if err != nil {
		// 不存在 → 创建
		if title == "" { title = "新对话" }
		now := time.Now()
		return db.Create(&AgentSession{
			SessionID: sessionID, UserID: userID,
			Title: truncateTitle(title, 32), Model: model,
			MessageCount: 1, LastMessageAt: &now,
		}).Error
	}
	// 存在 → 更新
	return db.Model(&existing).Updates(map[string]interface{}{
		"message_count":  existing.MessageCount + 1,
		"last_message_at": time.Now(),
	}).Error
}

// ListByUser 返回用户的所有会话，按最近活跃降序。
func (s *SessionStore) ListByUser(ctx context.Context, userID uint64) ([]AgentSession, error) {
	db := config.GetDB().WithContext(ctx)
	var sessions []AgentSession
	err := db.Where("user_id = ?", userID).Order("updated_at DESC").Limit(50).Find(&sessions).Error
	return sessions, err
}

// GetBySessionID 根据 session_id 查询单条会话。
func (s *SessionStore) GetBySessionID(ctx context.Context, sessionID string) (*AgentSession, error) {
	db := config.GetDB().WithContext(ctx)
	var session AgentSession
	err := db.Where("session_id = ?", sessionID).First(&session).Error
	if err != nil { return nil, err }
	return &session, nil
}

// Delete 删除会话（MySQL 元数据，Redis 消息由 SessionMemory 清理）。
func (s *SessionStore) Delete(ctx context.Context, sessionID string) error {
	db := config.GetDB().WithContext(ctx)
	return db.Where("session_id = ?", sessionID).Delete(&AgentSession{}).Error
}

// Rename 重命名会话。
func (s *SessionStore) Rename(ctx context.Context, sessionID, title string) error {
	db := config.GetDB().WithContext(ctx)
	return db.Model(&AgentSession{}).Where("session_id = ?", sessionID).Update("title", truncateTitle(title, 32)).Error
}

func truncateTitle(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen { return s }
	return string(runes[:maxLen]) + "…"
}
