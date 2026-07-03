package model

import (
	"time"
)

// BaseModel 基础模型
// 所有模型都继承自 BaseModel
type BaseModel struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"-"`
}

// User 用户模型
// 包含用户的基本信息和认证信息
type User struct {
	BaseModel
	Username     string `gorm:"size:64;uniqueIndex;not null" json:"username"`
	Password     string `gorm:"size:128;not null" json:"-"`
	Nickname     string `gorm:"size:64" json:"nickname"`
	Avatar       string  `gorm:"size:512" json:"avatar"`
	Phone        *string `gorm:"size:32;uniqueIndex" json:"phone"`
	Email        *string `gorm:"size:128;uniqueIndex" json:"email"`
	WechatOpenID *string `gorm:"size:128;uniqueIndex" json:"wechat_open_id"`
	Status       int    `gorm:"default:1" json:"status"`
}

// Route 路线模型
// 包含路线的基本信息和用户信息
type Route struct {
	BaseModel
	UserID         uint64       `gorm:"index;not null" json:"user_id"`
	Title          string       `gorm:"size:256;not null" json:"title"`
	Description    string       `gorm:"type:text" json:"description"`
	CoverImage     string       `gorm:"size:512" json:"cover_image"`
	StartCity      string       `gorm:"size:64" json:"start_city"`
	EndCity        string       `gorm:"size:64" json:"end_city"`
	TotalDistance  float64      `gorm:"type:decimal(10,2)" json:"total_distance"`
	EstimatedHours float64      `gorm:"type:decimal(10,2)" json:"estimated_hours"`
	PublishStatus  int          `gorm:"default:0" json:"publish_status"`
	ViewCount      int          `gorm:"default:0" json:"view_count"`
	LikeCount      int          `gorm:"default:0" json:"like_count"`
	FavoriteCount  int          `gorm:"default:0" json:"favorite_count"`
	ShareCount     int          `gorm:"default:0" json:"share_count"`
	ReuseCount     int          `gorm:"default:0" json:"reuse_count"`
	IsPublic       int          `gorm:"default:1" json:"is_public"`
	User           User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Checkpoints    []Checkpoint `gorm:"foreignKey:RouteID" json:"checkpoints,omitempty"`
}

// Checkpoint 路点模型
// 包含路线上的节点信息
type Checkpoint struct {
	BaseModel
	RouteID      uint64  `gorm:"index;not null" json:"route_id"`
	Name         string  `gorm:"size:128;not null" json:"name"`
	Description  string  `gorm:"type:text" json:"description"`
	Latitude     float64 `gorm:"type:decimal(10,7)" json:"latitude"`
	Longitude    float64 `gorm:"type:decimal(10,7)" json:"longitude"`
	Address      string  `gorm:"size:256" json:"address"`
	City         string  `gorm:"size:64" json:"city"`
	Sequence     int     `gorm:"not null" json:"sequence"`
	ArriveTime   string  `gorm:"size:32" json:"arrive_time"`
	StayDuration int     `gorm:"default:60" json:"stay_duration"`
	PhotoURL     string  `gorm:"size:512" json:"photo_url"`
	Route        Route   `gorm:"foreignKey:RouteID" json:"route,omitempty"`
}

// Checkin 打卡模型
// 包含用户在路线上的打卡记录
type Checkin struct {
	BaseModel
	UserID       uint64     `gorm:"index;not null" json:"user_id"`
	RouteID      uint64     `gorm:"index" json:"route_id"`
	CheckpointID uint64     `gorm:"index" json:"checkpoint_id"`
	CheckinTime  time.Time  `json:"checkin_time"`
	Latitude     float64    `gorm:"type:decimal(10,7)" json:"latitude"`
	Longitude    float64    `gorm:"type:decimal(10,7)" json:"longitude"`
	PhotoURL     string     `gorm:"size:512" json:"photo_url"`
	Content      string     `gorm:"type:text" json:"content"`
	Rating       int        `gorm:"default:5" json:"rating"`
	User         User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Checkpoint   Checkpoint `gorm:"foreignKey:CheckpointID" json:"checkpoint,omitempty"`
	Route        Route      `gorm:"foreignKey:RouteID" json:"route,omitempty"`
}

// Post 分享模型
// 包含用户在路线上的分享记录
type Post struct {
	BaseModel
	UserID       uint64 `gorm:"index;not null" json:"user_id"`
	RouteID      uint64 `gorm:"index" json:"route_id"`
	Title        string `gorm:"size:256;not null" json:"title"`
	Content      string `gorm:"type:text;not null" json:"content"`
	Images       string `gorm:"type:text" json:"images"`
	ViewCount    int    `gorm:"default:0" json:"view_count"`
	LikeCount    int    `gorm:"default:0" json:"like_count"`
	CommentCount int    `gorm:"default:0" json:"comment_count"`
	ShareCount   int    `gorm:"default:0" json:"share_count"`
	ReuseCount   int    `gorm:"default:0" json:"reuse_count"`
	Status       int    `gorm:"default:1" json:"status"`
	User         User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// Comment 评论模型
// 包含用户在分享上的评论记录
type Comment struct {
	BaseModel
	UserID    uint64 `gorm:"index;not null" json:"user_id"`
	PostID    uint64 `gorm:"index;not null" json:"post_id"`
	ParentID  uint64 `gorm:"index;default:0" json:"parent_id"`
	Content   string `gorm:"type:text;not null" json:"content"`
	LikeCount int    `gorm:"default:0" json:"like_count"`
	User      User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// Like 点赞模型
// 包含用户对分享的点赞记录
type Like struct {
	BaseModel
	UserID     uint64 `gorm:"uniqueIndex:idx_user_target;not null" json:"user_id"`
	TargetID   uint64 `gorm:"uniqueIndex:idx_user_target;not null" json:"target_id"`
	TargetType string `gorm:"size:32;not null" json:"target_type"`
}

// Favorite 收藏模型
// 包含用户对路线的收藏记录
type Favorite struct {
	BaseModel
	UserID  uint64 `gorm:"uniqueIndex:idx_user_favorite;not null" json:"user_id"`
	RouteID uint64 `gorm:"uniqueIndex:idx_user_favorite;not null" json:"route_id"`
}

// Share 分享模型
// 包含用户在分享上的分享记录
type Share struct {
	BaseModel
	UserID    uint64 `gorm:"index;not null" json:"user_id"`
	RouteID   uint64 `gorm:"index;not null" json:"route_id"`
	PostID    uint64 `gorm:"index" json:"post_id"`
	Platform  string `gorm:"size:32" json:"platform"`
	ShareCode string `gorm:"size:64;uniqueIndex" json:"share_code"`
}
