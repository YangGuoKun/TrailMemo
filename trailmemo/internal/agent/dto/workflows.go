package dto

// ── 游记 ──────────────────────────────────────────

// TravelNoteRequest 是游记生成的 API 请求体。
type TravelNoteRequest struct {
	RouteID              uint64 `json:"route_id" binding:"required"`
	Style                string `json:"style"` // story/journal/guide/social/poetic
	IncludeCheckinContent bool   `json:"include_checkin_content"`
	IncludeImages        bool   `json:"include_images"`
}

// TravelNoteResponse 是游记生成的 API 响应体。
type TravelNoteResponse struct {
	RunID            string   `json:"run_id"`
	ArtifactID       string   `json:"artifact_id"`
	Title            string   `json:"title"`
	Content          string   `json:"content"`
	SuggestedTags    []string `json:"suggested_tags"`
	ImageRefs        []string `json:"image_refs,omitempty"`
	Warnings         []string `json:"warnings,omitempty"`
	ApprovalRequired bool     `json:"approval_required"`
	NextAction       string   `json:"next_action"`
}

// ── 推荐 ──────────────────────────────────────────

// RecommendRequest 是推荐的 API 请求体。
type RecommendRequest struct {
	Query      string   `json:"query" binding:"required"`
	Days       int      `json:"days"`
	Budget     int      `json:"budget"`
	Interests  []string `json:"interests"`
	TravelType string   `json:"travel_type"`
}

// RecommendItem 是一条推荐条目。
type RecommendItem struct {
	Title           string   `json:"title"`
	City            string   `json:"city"`
	Reason          string   `json:"reason"`
	EstimatedBudget int      `json:"estimated_budget"`
	Days            int      `json:"days"`
	Tags            []string `json:"tags"`
}

// RecommendResponse 是推荐的 API 响应体。
type RecommendResponse struct {
	RunID      string          `json:"run_id"`
	ArtifactID string          `json:"artifact_id"`
	Items      []RecommendItem `json:"items"`
	Fallback   bool            `json:"fallback"`
	Warnings   []string        `json:"warnings,omitempty"`
}

// ── 复用改造 ──────────────────────────────────────

// RemixRequest 是路线复用改造的 API 请求体。
type RemixRequest struct {
	Query        string   `json:"query" binding:"required"`
	Days         int      `json:"days"`
	Budget       int      `json:"budget"`
	TravelStyles []string `json:"travel_styles"`
}

// RemixResponse 是路线复用改造的 API 响应体。
type RemixResponse struct {
	RunID          string              `json:"run_id"`
	SourceRouteID  uint64              `json:"source_route_id"`
	ArtifactID     string              `json:"artifact_id"`
	ChangeSummary  []RemixChangeItem   `json:"change_summary"`
	RouteDraft     *RouteDraftData     `json:"route_draft"`
	Warnings       []string            `json:"warnings,omitempty"`
	ApprovalRequired bool              `json:"approval_required"`
	NextAction     string              `json:"next_action"`
}

// RemixChangeItem 是改造变更说明。
type RemixChangeItem struct {
	Action string `json:"action"` // added/removed/modified
	Point  string `json:"point"`  // 打卡点名称
	Reason string `json:"reason"` // 变更原因
}
