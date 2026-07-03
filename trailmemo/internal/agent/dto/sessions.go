package dto

// SessionInfo 是会话列表中的单条会话摘要。
type SessionInfo struct {
	SessionID     string `json:"session_id"`
	Title         string `json:"title"`
	Model         string `json:"model"`
	MessageCount  int    `json:"message_count"`
	LastMessageAt string `json:"last_message_at,omitempty"`
	CreatedAt     string `json:"created_at"`
}

// SessionListResponse 是会话列表 API 响应。
type SessionListResponse struct {
	Sessions []SessionInfo `json:"sessions"`
	Total    int           `json:"total"`
}

// SessionDetailResponse 是会话详情 API 响应（含消息历史）。
type SessionDetailResponse struct {
	SessionID    string          `json:"session_id"`
	Title        string          `json:"title"`
	Messages     []SessionMsg    `json:"messages"`
	MessageCount int             `json:"message_count"`
	Expired      bool            `json:"expired"`
}

// SessionMsg 是一条历史消息。
type SessionMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// RenameSessionRequest 是重命名会话的请求体。
type RenameSessionRequest struct {
	Title string `json:"title" binding:"required"`
}
