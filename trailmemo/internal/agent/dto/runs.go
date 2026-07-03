package dto

type RunDetailResponse struct {
	RunID       string             `json:"run_id"`
	UserID      uint64             `json:"user_id"`
	SessionID   string             `json:"session_id,omitempty"`
	Intent      string             `json:"intent"`
	Mode        string             `json:"mode"`
	Status      string             `json:"status"`
	Model       string             `json:"model"`
	PromptVer   string             `json:"prompt_version,omitempty"`
	TotalTokens int                `json:"total_tokens"`
	LatencyMs   int64              `json:"latency_ms"`
	ErrorCode   string             `json:"error_code,omitempty"`
	ErrorMsg    string             `json:"error_message,omitempty"`
	CreatedAt   string             `json:"created_at"`
	UpdatedAt   string             `json:"updated_at"`
	Steps       []RunStepInfo      `json:"steps"`
	Artifacts   []RunArtifactInfo  `json:"artifacts"`
}

type RunStepInfo struct {
	Index     int    `json:"index"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	LatencyMs int64  `json:"latency_ms"`
	CreatedAt string `json:"created_at"`
}

type RunArtifactInfo struct {
	ArtifactID          string  `json:"artifact_id"`
	Type                string  `json:"type"`
	Status              string  `json:"status"`
	CommittedEntityType string  `json:"committed_entity_type,omitempty"`
	CommittedEntityID   *uint64 `json:"committed_entity_id,omitempty"`
	CreatedAt           string  `json:"created_at"`
}
