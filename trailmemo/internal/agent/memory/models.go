// Package memory 提供 Agent 运行记录、步骤、产物和用户画像的持久化存储。
// 严格按照 AI_AGENT_GO_ARCHITECTURE_DESIGN.md §11 数据模型设计实现。
package memory

import (
	"time"
)

// AgentRun 记录一次 Agent 任务的完整生命周期。
// 对应设计文档 §11.2 agent_runs 表。
type AgentRun struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement" json:"id"`                         // 自增主键
	RunID        string     `gorm:"size:64;uniqueIndex;not null" json:"run_id"`                 // 外部可见运行 ID
	UserID       uint64     `gorm:"index;not null" json:"user_id"`                              // 发起用户 ID
	SessionID    string     `gorm:"size:64;index" json:"session_id"`                            // 会话 ID，关联多轮对话
	Intent       string     `gorm:"size:64" json:"intent"`                                      // 意图：route_draft/recommend/travel_note/chat
	Mode         string     `gorm:"size:32" json:"mode"`                                        // 执行模式：one_shot/workflow/agent_loop
	Status       string     `gorm:"size:32;default:created" json:"status"`                      // 状态：created/running/completed/failed/cancelled
	InputSummary string     `gorm:"type:text" json:"input_summary"`                             // 脱敏后的输入摘要
	Model        string     `gorm:"size:64" json:"model"`                                       // 使用的模型名
	PromptVer    string     `gorm:"size:64" json:"prompt_version"`                              // prompt 模板版本
	TotalTokens  int        `gorm:"default:0" json:"total_tokens"`                              // token 总量
	LatencyMs    int64      `gorm:"default:0" json:"latency_ms"`                                // 总耗时（毫秒）
	ErrorCode    string     `gorm:"size:64" json:"error_code"`                                  // 错误码
	ErrorMsg     string     `gorm:"type:text" json:"error_message"`                             // 脱敏错误信息
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`                           // 创建时间
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`                           // 更新时间
}

func (AgentRun) TableName() string { return "agent_runs" }

// AgentStep 记录 Agent 运行中的单个步骤。
// 对应设计文档 §11.3 agent_steps 表。
type AgentStep struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`                             // 自增主键
	RunID     string    `gorm:"size:64;index;not null" json:"run_id"`                           // 关联的 run_id
	StepIdx   int       `gorm:"not null" json:"step_index"`                                     // 步骤序号
	StepType  string    `gorm:"size:32" json:"step_type"`                                       // 步骤类型：llm/tool/validation/approval/fallback
	Name      string    `gorm:"size:128" json:"name"`                                           // 步骤名称
	Status    string    `gorm:"size:32" json:"status"`                                          // 状态：running/success/failed/skipped
	InputJSON string    `gorm:"type:json" json:"input_json"`                                    // 脱敏输入
	OutputJSON string   `gorm:"type:json" json:"output_json"`                                   // 脱敏输出
	LatencyMs int64     `gorm:"default:0" json:"latency_ms"`                                    // 耗时（毫秒）
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`                               // 创建时间
}

func (AgentStep) TableName() string { return "agent_steps" }

// AgentArtifact 记录 Agent 生成的产物（路线草稿、游记草稿、推荐结果）。
// 对应设计文档 §11.4 agent_artifacts 表。
type AgentArtifact struct {
	ID                 uint64     `gorm:"primaryKey;autoIncrement" json:"id"`                    // 自增主键
	ArtifactID         string     `gorm:"size:64;uniqueIndex;not null" json:"artifact_id"`       // 产物唯一 ID
	RunID              string     `gorm:"size:64;index;not null" json:"run_id"`                  // 关联的 run_id
	UserID             uint64     `gorm:"index;not null" json:"user_id"`                         // 用户 ID
	Type               string     `gorm:"size:64" json:"type"`                                   // 产物类型：route_draft/travel_note/recommendation
	Status             string     `gorm:"size:32;default:draft" json:"status"`                   // 状态：draft/committed/expired/cancelled
	ContentJSON        string     `gorm:"type:json" json:"content_json"`                         // 结构化产物内容
	CommittedEntityType string    `gorm:"size:64" json:"committed_entity_type"`                  // 提交后的业务实体类型：route/post
	CommittedEntityID  *uint64    `json:"committed_entity_id"`                                   // 提交后的业务实体 ID
	ExpiresAt          *time.Time `json:"expires_at"`                                            // 草稿过期时间
	CreatedAt          time.Time  `gorm:"autoCreateTime" json:"created_at"`                      // 创建时间
	UpdatedAt          time.Time  `gorm:"autoUpdateTime" json:"updated_at"`                      // 更新时间
}

func (AgentArtifact) TableName() string { return "agent_artifacts" }

// AgentUserPreference 存储 AI 用户画像。
// 对应设计文档 §11.5 agent_user_preferences 表。
type AgentUserPreference struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement" json:"id"`                        // 自增主键
	UserID          uint64    `gorm:"uniqueIndex;not null" json:"user_id"`                       // 用户 ID
	TravelStyles    string    `gorm:"type:json" json:"travel_styles"`                            // 旅行风格 JSON 数组
	BudgetLevel     string    `gorm:"size:32" json:"budget_level"`                               // 预算等级：low/medium/high/custom
	Pace            string    `gorm:"size:32" json:"pace"`                                       // 节奏偏好：slow/normal/intense
	CompanionTypes  string    `gorm:"type:json" json:"companion_types"`                          // 出行伙伴 JSON 数组
	Interests       string    `gorm:"type:json" json:"interests"`                                // 兴趣标签 JSON 数组
	PreferredCities string    `gorm:"type:json" json:"preferred_cities"`                         // 偏好城市 JSON 数组
	AvoidedCities   string    `gorm:"type:json" json:"avoided_cities"`                           // 避免城市 JSON 数组
	DislikedFactors string    `gorm:"type:json" json:"disliked_factors"`                         // 不喜欢因素 JSON 数组
	Confidence      float64   `gorm:"type:decimal(4,3);default:0" json:"confidence"`             // 画像置信度 0-1
	Source          string    `gorm:"size:32;default:explicit" json:"source"`                    // 来源：explicit/behavior/mixed
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`                          // 创建时间
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`                          // 更新时间
}

func (AgentUserPreference) TableName() string { return "agent_user_preferences" }

// AgentSession 记录一次用户对话会话的元数据。
// 消息内容存 Redis，此表只存标题/计数/时间用于列表展示。
type AgentSession struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	SessionID     string     `gorm:"size:64;uniqueIndex;not null" json:"session_id"`
	UserID        uint64     `gorm:"index;not null" json:"user_id"`
	Title         string     `gorm:"size:128;not null;default:新对话" json:"title"`
	Model         string     `gorm:"size:64" json:"model"`
	MessageCount  int        `gorm:"default:0" json:"message_count"`
	LastMessageAt *time.Time `json:"last_message_at"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (AgentSession) TableName() string { return "agent_sessions" }
