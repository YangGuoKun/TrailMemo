package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/trailmemo/internal/config"
	"github.com/trailmemo/internal/platform/logger"
	"go.uber.org/zap"
)

// SessionMessage 是一条会话消息。
type SessionMessage struct {
	Role    string `json:"role"`    // system/user/assistant/tool
	Content string `json:"content"` // 消息内容
}

// SessionMemory 管理 Agent 多轮对话的短期记忆。
// Redis 存储，TTL 按配置，支持追加消息和检索最近 N 轮。
type SessionMemory struct {
	ttl time.Duration
}

func NewSessionMemory(ttl time.Duration) *SessionMemory {
	return &SessionMemory{ttl: ttl}
}

// sessionKey 生成 Redis key。
func (s *SessionMemory) sessionKey(sessionID string) string {
	return fmt.Sprintf("agent:session:%s", sessionID)
}

// AppendMessages 向会话追加消息。仅保留最近 maxMessages 条。
func (s *SessionMemory) AppendMessages(ctx context.Context, sessionID string, msgs []SessionMessage, maxMessages int) error {
	rdb := config.GetRedis()
	if rdb == nil {
		return nil // Redis 不可用时静默跳过
	}

	// 读取现有消息
	existing, _ := s.GetMessages(ctx, sessionID, 100)
	existing = append(existing, msgs...)

	// 只保留最近 maxMessages 条
	if len(existing) > maxMessages {
		existing = existing[len(existing)-maxMessages:]
	}

	data, err := json.Marshal(existing)
	if err != nil {
		return err
	}

	key := s.sessionKey(sessionID)
	if err := rdb.Set(ctx, key, data, s.ttl).Err(); err != nil {
		logger.FromContext(ctx).Warn("session_memory_save_failed",
			zap.Error(err), zap.String("session_id", sessionID))
		return fmt.Errorf("保存会话失败: %w", err)
	}
	return nil
}

// GetMessages 获取会话的全部消息。
func (s *SessionMemory) GetMessages(ctx context.Context, sessionID string, maxMessages int) ([]SessionMessage, error) {
	rdb := config.GetRedis()
	if rdb == nil {
		return nil, nil
	}

	key := s.sessionKey(sessionID)
	data, err := rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil, nil // key 不存在视为空会话
	}

	var msgs []SessionMessage
	if err := json.Unmarshal(data, &msgs); err != nil {
		return nil, nil
	}

	if len(msgs) > maxMessages {
		msgs = msgs[len(msgs)-maxMessages:]
	}
	return msgs, nil
}

// ClearSession 清除指定会话。
func (s *SessionMemory) ClearSession(ctx context.Context, sessionID string) error {
	rdb := config.GetRedis()
	if rdb == nil {
		return nil
	}
	return rdb.Del(ctx, s.sessionKey(sessionID)).Err()
}
