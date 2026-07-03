// Package observability 提供 Agent 运行时的可观测性指标。
// Phase 6 使用内存计数器 + zap 日志输出；Phase 7+ 可接入 Prometheus / OpenTelemetry。
package observability

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/trailmemo/internal/platform/logger"
	"go.uber.org/zap"
)

// Metrics 是 Agent 模块的核心指标汇总，线程安全。
type Metrics struct {
	mu sync.RWMutex

	// 运行统计
	RunTotal        atomic.Int64 // Agent 请求总数
	RunSuccess      atomic.Int64 // 成功数
	RunFailed       atomic.Int64 // 失败数
	RunCancelled    atomic.Int64 // 取消数

	// 按意图分类
	RunByIntent map[string]*atomic.Int64 // intent → count

	// 工具调用
	ToolCallTotal atomic.Int64
	ToolCallFail  atomic.Int64

	// LLM
	LLMCallTotal   atomic.Int64
	LLMTokenTotal  atomic.Int64 // 累计 token
	LLMRetryTotal  atomic.Int64 // 重试次数
	FallbackTotal  atomic.Int64 // 降级次数

	// 延迟
	TotalLatencyMs atomic.Int64 // 累计延迟
	MaxLatencyMs   atomic.Int64 // 最大延迟

	// 安全
	GuardrailBlocked atomic.Int64 // 输入拦截次数
	ApprovalPending  atomic.Int64 // 待审批产物数

	startTime time.Time
}

var globalMetrics = &Metrics{
	RunByIntent: make(map[string]*atomic.Int64),
	startTime:   time.Now(),
}

// GetMetrics 返回全局指标单例。
func GetMetrics() *Metrics { return globalMetrics }

// RecordRun 记录一次 Agent 运行的完成情况。
func (m *Metrics) RecordRun(intent string, success bool, latencyMs int64, tokens int) {
	m.RunTotal.Add(1)
	if success {
		m.RunSuccess.Add(1)
	} else {
		m.RunFailed.Add(1)
	}
	m.TotalLatencyMs.Add(latencyMs)
	if latencyMs > m.MaxLatencyMs.Load() {
		m.MaxLatencyMs.Store(latencyMs)
	}
	m.LLMTokenTotal.Add(int64(tokens))
	m.getIntentCounter(intent).Add(1)
}

// RecordToolCall 记录工具调用结果。
func (m *Metrics) RecordToolCall(success bool) {
	m.ToolCallTotal.Add(1)
	if !success { m.ToolCallFail.Add(1) }
}

// RecordLLMCall 记录 LLM 调用统计。
func (m *Metrics) RecordLLMCall(tokens int) {
	m.LLMCallTotal.Add(1)
	m.LLMTokenTotal.Add(int64(tokens))
}

// RecordRetry 记录 LLM 重试。
func (m *Metrics) RecordRetry() { m.LLMRetryTotal.Add(1) }

// RecordFallback 记录降级触发。
func (m *Metrics) RecordFallback() { m.FallbackTotal.Add(1) }

// RecordGuardrailBlock 记录安全拦截。
func (m *Metrics) RecordGuardrailBlock() { m.GuardrailBlocked.Add(1) }

func (m *Metrics) getIntentCounter(intent string) *atomic.Int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c, ok := m.RunByIntent[intent]; ok { return c }
	c := &atomic.Int64{}
	m.RunByIntent[intent] = c
	return c
}

// Snapshot 返回当前指标的不可变快照。
func (m *Metrics) Snapshot() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	intents := make(map[string]int64)
	for k, v := range m.RunByIntent { intents[k] = v.Load() }
	avgLat := int64(0)
	if m.RunTotal.Load() > 0 {
		avgLat = m.TotalLatencyMs.Load() / m.RunTotal.Load()
	}
	return MetricsSnapshot{
		RunTotal:        m.RunTotal.Load(),
		RunSuccess:      m.RunSuccess.Load(),
		RunFailed:       m.RunFailed.Load(),
		RunByIntent:     intents,
		ToolCallTotal:   m.ToolCallTotal.Load(),
		ToolCallFail:    m.ToolCallFail.Load(),
		LLMCallTotal:    m.LLMCallTotal.Load(),
		LLMTokenTotal:   m.LLMTokenTotal.Load(),
		LLMRetryTotal:   m.LLMRetryTotal.Load(),
		FallbackTotal:   m.FallbackTotal.Load(),
		AvgLatencyMs:    avgLat,
		MaxLatencyMs:    m.MaxLatencyMs.Load(),
		GuardrailBlocked: m.GuardrailBlocked.Load(),
		UptimeSeconds:   int64(time.Since(m.startTime).Seconds()),
	}
}

// LogSummary 定期输出指标摘要到日志。
func (m *Metrics) LogSummary() {
	s := m.Snapshot()
	logger.L().Info("agent_metrics_summary",
		zap.Int64("run_total", s.RunTotal),
		zap.Int64("run_success", s.RunSuccess),
		zap.Int64("run_failed", s.RunFailed),
		zap.Int64("llm_call_total", s.LLMCallTotal),
		zap.Int64("llm_token_total", s.LLMTokenTotal),
		zap.Int64("fallback_total", s.FallbackTotal),
		zap.Int64("avg_latency_ms", s.AvgLatencyMs),
		zap.Int64("uptime_s", s.UptimeSeconds),
	)
}

// MetricsSnapshot 是指标的只读快照，用于 API 响应。
type MetricsSnapshot struct {
	RunTotal        int64            `json:"run_total"`
	RunSuccess      int64            `json:"run_success"`
	RunFailed       int64            `json:"run_failed"`
	RunByIntent     map[string]int64 `json:"run_by_intent"`
	ToolCallTotal   int64            `json:"tool_call_total"`
	ToolCallFail    int64            `json:"tool_call_fail"`
	LLMCallTotal    int64            `json:"llm_call_total"`
	LLMTokenTotal   int64            `json:"llm_token_total"`
	LLMRetryTotal   int64            `json:"llm_retry_total"`
	FallbackTotal   int64            `json:"fallback_total"`
	AvgLatencyMs    int64            `json:"avg_latency_ms"`
	MaxLatencyMs    int64            `json:"max_latency_ms"`
	GuardrailBlocked int64           `json:"guardrail_blocked"`
	UptimeSeconds   int64            `json:"uptime_seconds"`
}
