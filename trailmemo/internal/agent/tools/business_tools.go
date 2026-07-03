package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/trailmemo/internal/service"
)

// ── RouteTool ────────────────────────────────────

// RouteTool 提供路线相关查询能力。
// 只读操作，权限等级 PermissionRead。
type RouteTool struct {
	routeService service.RouteService
}

func NewRouteTool(rs service.RouteService) *RouteTool {
	return &RouteTool{routeService: rs}
}

func (t *RouteTool) Name() string        { return "route.search_public" }
func (t *RouteTool) Description() string { return "查询公开路线，作为推荐候选。可指定城市或关键词筛选。" }
func (t *RouteTool) Permission() Permission { return PermissionRead }
func (t *RouteTool) JSONSchema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"city":{"type":"string","description":"筛选城市"},"keyword":{"type":"string","description":"搜索关键词"},"limit":{"type":"integer","description":"最多返回条数"}}}`)
}

// routeSearchArgs 是 route.search_public 的参数结构。
type routeSearchArgs struct {
	City    string `json:"city"`
	Keyword string `json:"keyword"`
	Limit   int    `json:"limit"`
}

func (t *RouteTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
	var params routeSearchArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return &ToolResult{Success: false, Error: "参数解析失败"}, err
	}
	if params.Limit <= 0 {
		params.Limit = 5
	}

	// 调用现有 Service 接口（按公开路线分页查询，后续可加城市/关键词筛选）
	routes, total, err := t.routeService.GetPublicRoutes(ctx, 1, params.Limit)
	if err != nil {
		return &ToolResult{Success: false, Error: fmt.Sprintf("查询路线失败: %v", err)}, err
	}

	return &ToolResult{Success: true, Data: map[string]interface{}{
		"routes": routes,
		"total":  total,
	}}, nil
}

// ── CheckinTool ──────────────────────────────────

// CheckinTool 提供打卡记录查询能力。
type CheckinTool struct {
	checkinService service.CheckinService
}

func NewCheckinTool(cs service.CheckinService) *CheckinTool {
	return &CheckinTool{checkinService: cs}
}

func (t *CheckinTool) Name() string            { return "checkin.list_by_route" }
func (t *CheckinTool) Description() string     { return "查询指定路线的打卡记录，包含照片、评分、感受等内容。" }
func (t *CheckinTool) Permission() Permission { return PermissionRead }
func (t *CheckinTool) JSONSchema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"route_id":{"type":"integer","description":"路线ID"},"page":{"type":"integer","description":"页码"},"size":{"type":"integer","description":"每页数量"}},"required":["route_id"]}`)
}

type checkinListArgs struct {
	RouteID uint64 `json:"route_id"`
	Page    int    `json:"page"`
	Size    int    `json:"size"`
}

func (t *CheckinTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
	var params checkinListArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return &ToolResult{Success: false, Error: "参数解析失败"}, err
	}
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Size <= 0 {
		params.Size = 20
	}
	checkins, total, err := t.checkinService.GetCheckinsByRouteID(ctx, params.RouteID, params.Page, params.Size)
	if err != nil {
		return &ToolResult{Success: false, Error: fmt.Sprintf("查询打卡失败: %v", err)}, err
	}
	return &ToolResult{Success: true, Data: map[string]interface{}{
		"checkins": checkins,
		"total":    total,
	}}, nil
}

// ── CommunityTool ────────────────────────────────

// CommunityTool 提供社区内容查询能力。
type CommunityTool struct {
	postService service.PostService
}

func NewCommunityTool(ps service.PostService) *CommunityTool {
	return &CommunityTool{postService: ps}
}

func (t *CommunityTool) Name() string            { return "community.search_posts" }
func (t *CommunityTool) Description() string     { return "搜索社区公开帖子，按关键词或城市筛选，用于推荐参考。" }
func (t *CommunityTool) Permission() Permission { return PermissionRead }
func (t *CommunityTool) JSONSchema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"keyword":{"type":"string","description":"搜索关键词"},"city":{"type":"string","description":"筛选城市"},"limit":{"type":"integer","description":"最多返回条数"}}}`)
}

type communitySearchArgs struct {
	Keyword string `json:"keyword"`
	City    string `json:"city"`
	Limit   int    `json:"limit"`
}

func (t *CommunityTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
	var params communitySearchArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return &ToolResult{Success: false, Error: "参数解析失败"}, err
	}
	if params.Limit <= 0 {
		params.Limit = 5
	}

	// 使用现有 GetAllPublicPosts 接口，后续可扩展关键词/城市搜索
	posts, total, err := t.postService.GetAllPublicPosts(ctx, 1, params.Limit)
	if err != nil {
		return &ToolResult{Success: false, Error: fmt.Sprintf("搜索帖子失败: %v", err)}, err
	}
	return &ToolResult{Success: true, Data: map[string]interface{}{
		"posts": posts,
		"total":  total,
	}}, nil
}
