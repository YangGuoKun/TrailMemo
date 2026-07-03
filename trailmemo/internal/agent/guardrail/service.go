// Package guardrail 提供 Agent 安全护栏：输入过滤、输出校验、审批门禁、幂等存储。
package guardrail

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/trailmemo/internal/config"
	"github.com/trailmemo/internal/platform/logger"
	"go.uber.org/zap"
)

// containsWord 检查输入中是否包含指定关键词。
func containsWord(input, keyword string) bool {
	return strings.Contains(strings.ToLower(input), strings.ToLower(keyword))
}

// Service 是 Guardrail 的统一入口。
type Service struct {
	cfg       config.AgentConfig
	validator *Validator
}

// NewService 创建 Guardrail 服务实例。
func NewService(cfg config.AgentConfig) *Service {
	return &Service{
		cfg:       cfg,
		validator: NewValidator(1),
	}
}

// CheckInput 对用户输入做基本安全检查。
// 返回 nil 表示通过，否则返回拒绝原因。
func (s *Service) CheckInput(ctx context.Context, userInput string) error {
	if len(userInput) == 0 {
		return fmt.Errorf("输入为空")
	}
	maxLen := s.cfg.MaxInputLength
	if maxLen == 0 {
		maxLen = 5000
	}
	if len(userInput) > maxLen {
		return fmt.Errorf("输入过长（最多%d字）", maxLen)
	}
	// 检测危险操作关键词
	dangerous := []string{"删除所有", "清空数据库", "drop table", "rm -rf"}
	for _, kw := range dangerous {
		if containsWord(userInput, kw) {
			logger.FromContext(ctx).Warn("guardrail_blocked_dangerous_input",
				zap.String("keyword", kw))
			return fmt.Errorf("输入包含不被允许的操作")
		}
	}
	return nil
}

// ValidateArtifactOutput 对 LLM 生成的输出做 JSON 校验。
func (s *Service) ValidateArtifactOutput(raw string) (*ValidationResult, error) {
	return s.validator.ValidateJSON(raw)
}

// NeedsApproval 判断指定操作是否需要用户确认。
// 对应 ADR-004：真正的写操作和公开操作必须确认。
func (s *Service) NeedsApproval(commitType string) bool {
	switch commitType {
	case "create_route", "create_post":
		return true
	case "recommendation":
		return false
	default:
		return s.cfg.Approval.RequireForUserWrite
	}
}

// ── 幂等存储 ──────────────────────────────────────────

// CheckIdempotency 检查 idempotency_key 是否已使用。
// 返回已存在的结果，或 nil 表示是新请求。
func (s *Service) CheckIdempotency(ctx context.Context, userID uint64, key string) (string, error) {
	rdb := config.GetRedis()
	if rdb == nil {
		return "", nil // Redis 不可用时跳过幂等检查
	}
	redisKey := fmt.Sprintf("agent:idempotency:%d:%s", userID, key)
	val, err := rdb.Get(ctx, redisKey).Result()
	if err != nil {
		return "", nil // key 不存在
	}
	return val, nil
}

// SaveIdempotency 保存幂等键的结果。
func (s *Service) SaveIdempotency(ctx context.Context, userID uint64, key, result string) error {
	rdb := config.GetRedis()
	if rdb == nil {
		return nil
	}
	redisKey := fmt.Sprintf("agent:idempotency:%d:%s", userID, key)
	return rdb.Set(ctx, redisKey, result, 24*time.Hour).Err()
}
