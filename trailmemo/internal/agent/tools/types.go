package tools

// Permission defines the approval boundary for an agent tool.
type Permission string

const (
	PermissionRead         Permission = "read"          // 读取权限
	PermissionDraftWrite   Permission = "draft_write"   // 草稿写入权限
	PermissionUserWrite    Permission = "user_write"    // 用户写入权限
	PermissionPublicAction Permission = "public_action" // 公共操作权限
	PermissionDangerous    Permission = "dangerous"     // 危险操作权限
)

// ToolDescriptor is the stable description exposed by the registry.
type ToolDescriptor struct {
	Name        string     `json:"name"`        // 工具名称
	Description string     `json:"description"` // 工具描述
	Permission  Permission `json:"permission"`  // 工具权限
	Phase       string     `json:"phase"`       // 工具阶段
	Enabled     bool       `json:"enabled"`     // 是否启用
}

// BootstrapToolDescriptors returns the planned initial tool surface without enabling execution yet.
// 该函数用于在系统启动时初始化工具描述符，但不启用执行。
func BootstrapToolDescriptors() []ToolDescriptor {
	return []ToolDescriptor{
		{
			Name:        "route.search_public",
			Description: "Search public routes as recommendation candidates.",
			Permission:  PermissionRead,
			Phase:       "P2",
			Enabled:     false,
		},
		{
			Name:        "route.get_detail",
			Description: "Read a route and its checkpoints through RouteService.",
			Permission:  PermissionRead,
			Phase:       "P2",
			Enabled:     false,
		},
		{
			Name:        "route.create_from_artifact",
			Description: "Create a user route from an approved agent route draft artifact.",
			Permission:  PermissionUserWrite,
			Phase:       "P3",
			Enabled:     false,
		},
		{
			Name:        "checkin.list_by_route",
			Description: "Read checkins for a route through CheckinService.",
			Permission:  PermissionRead,
			Phase:       "P2",
			Enabled:     false,
		},
		{
			Name:        "post.create_from_artifact",
			Description: "Publish an approved travel note artifact as a community post.",
			Permission:  PermissionPublicAction,
			Phase:       "P3",
			Enabled:     false,
		},
	}
}
