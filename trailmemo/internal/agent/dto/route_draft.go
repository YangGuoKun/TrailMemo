package dto

// RouteDraftRequest 是路线草稿生成的 API 请求体。
type RouteDraftRequest struct {
	SessionID    string   `json:"session_id"`               // 会话 ID
	Query        string   `json:"query" binding:"required"` // 用户自然语言需求
	StartCity    string   `json:"start_city"`               // 出发城市
	TargetCity   string   `json:"target_city"`              // 目标城市
	Days         int      `json:"days"`                     // 出行天数
	Budget       int      `json:"budget"`                   // 预算（元）
	TravelStyles []string `json:"travel_styles"`            // 旅行风格标签
}

// RouteDraftResponse 是路线草稿生成的 API 响应体。
type RouteDraftResponse struct {
	RunID            string          `json:"run_id"`             // Agent 运行 ID
	ArtifactID       string          `json:"artifact_id"`        // 产物 ID
	RouteDraft       *RouteDraftData `json:"route_draft"`        // 路线草稿数据
	Warnings         []string        `json:"warnings,omitempty"` // 警告
	ApprovalRequired bool            `json:"approval_required"`  // 是否需要确认
	NextAction       string          `json:"next_action"`        // 后续操作
}

// RouteDraftData 是路线草稿的对外数据结构。
type RouteDraftData struct {
	Title           string                `json:"title"`
	Summary         string                `json:"summary"`
	StartCity       string                `json:"start_city"`
	EndCity         string                `json:"end_city"`
	EstimatedBudget float64               `json:"estimated_budget"`
	EstimatedHours  float64               `json:"estimated_hours"`
	Checkpoints     []CheckpointDraftData `json:"checkpoints"`
}

// CheckpointDraftData 是打卡点草稿的对外数据结构。
type CheckpointDraftData struct {
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	City         string  `json:"city"`
	Address      string  `json:"address"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Sequence     int     `json:"sequence"`
	ArriveTime   string  `json:"arrive_time"`
	StayDuration int     `json:"stay_duration"`
}

// ArtifactCommitRequest 是产物提交的 API 请求体。
type ArtifactCommitRequest struct {
	CommitType     string `json:"commit_type" binding:"required"`     // "create_route" / "create_post"
	IdempotencyKey string `json:"idempotency_key" binding:"required"` // 幂等键
	IsPublic       int    `json:"is_public"`                          // 是否公开（仅 create_route）
}

// ArtifactCommitResponse 是产物提交的 API 响应体。
type ArtifactCommitResponse struct {
	ArtifactID string `json:"artifact_id"` // 产物 ID
	Status     string `json:"status"`      // committed
	EntityType string `json:"entity_type"` // route / post
	EntityID   uint64 `json:"entity_id"`   // 业务实体 ID
}

// ArtifactApprovalResponse 是产物确认 API 响应。
type ArtifactApprovalResponse struct {
	ArtifactID string `json:"artifact_id"`
	Status     string `json:"status"`
}
