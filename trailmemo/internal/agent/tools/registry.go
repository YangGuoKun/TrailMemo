// Package tools 实现 Agent 工具注册中心和执行器。
// 严格按照 ADR-003：工具只调用 Service 层，不直接访问 Repository。
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/trailmemo/internal/platform/logger"
	"go.uber.org/zap"
)

// Tool 接口定义了 Agent 可调用的能力单元。
// 对应设计文档 §8.1 Tool 接口边界。
type Tool interface {
	// Name 返回工具唯一名称，如 "route.search_public"。
	Name() string
	// Description 返回给模型看的能力说明，必须准确描述边界。
	Description() string
	// Permission 返回工具权限等级：read/draft_write/user_write/public_action/dangerous。
	Permission() Permission
	// JSONSchema 返回参数 schema，用于模型 tool call 和服务端校验。
	JSONSchema() json.RawMessage
	// Execute 执行业务逻辑。只能调用 service 层，禁止绕过业务规则直接写库。
	Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error)
}

// ToolResult 是工具执行的统一返回结构。
type ToolResult struct {
	Success bool        `json:"success"`         // 是否成功
	Data    interface{} `json:"data,omitempty"`  // 成功时的结构化数据
	Error   string      `json:"error,omitempty"` // 失败时的错误信息
}

// Registry 是工具注册中心，管理所有已注册工具的发现和执行。
// 线程安全，支持动态注册（未来可接 MCP）。
type Registry struct {
	mu    sync.RWMutex
	tools map[string]Tool // name → tool
}

// NewRegistry 创建一个空注册中心。
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register 注册一个工具。同名工具后注册的会覆盖先注册的。
func (r *Registry) Register(tool Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Name()] = tool
	logger.L().Info("tool_registered",
		zap.String("name", tool.Name()),
		zap.String("permission", string(tool.Permission())))
}

// Get 根据名称获取工具实例。
func (r *Registry) Get(name string) (Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("工具未注册: %s", name)
	}
	return t, nil
}

// Execute 根据名称和参数执行一个工具。
// 同时校验工具权限，拒绝执行 dangerous 级别工具（除非显式允许）。
func (r *Registry) Execute(ctx context.Context, name string, args json.RawMessage) (*ToolResult, error) {
	tool, err := r.Get(name)
	if err != nil {
		return nil, err
	}

	// 安全检查：危险操作默认拒绝
	if tool.Permission() == PermissionDangerous {
		return &ToolResult{Success: false, Error: "危险操作需要管理确认"}, nil
	}
	if err := validateToolArgs(tool.JSONSchema(), args); err != nil {
		logger.FromContext(ctx).Warn("tool_args_validation_failed",
			zap.String("tool", name), zap.Error(err))
		return &ToolResult{Success: false, Error: err.Error()}, nil
	}

	result, err := tool.Execute(ctx, args)
	if err != nil {
		logger.FromContext(ctx).Error("tool_execute_failed",
			zap.String("tool", name), zap.Error(err))
	} else {
		logger.FromContext(ctx).Info("tool_execute_success",
			zap.String("tool", name))
	}
	return result, err
}

// GetAllDescriptors 返回所有已注册工具的描述符列表，用于 /capabilities 接口。
func (r *Registry) GetAllDescriptors() []ToolDescriptor {
	r.mu.RLock()
	defer r.mu.RUnlock()
	descs := make([]ToolDescriptor, 0, len(r.tools))
	for _, t := range r.tools {
		descs = append(descs, ToolDescriptor{
			Name:        t.Name(),
			Description: t.Description(),
			Permission:  t.Permission(),
			Enabled:     true,
			Phase:       "P2",
		})
	}
	return descs
}

// GetToolDefs 返回所有工具的函数定义列表，用于传给 LLM 的 tools 参数。
func (r *Registry) GetToolDefs() []ToolDef {
	r.mu.RLock()
	defer r.mu.RUnlock()
	defs := make([]ToolDef, 0, len(r.tools))
	for _, t := range r.tools {
		defs = append(defs, ToolDef{
			Function: FunctionDef{
				Name:        t.Name(),
				Description: t.Description(),
				Parameters:  t.JSONSchema(),
			},
		})
	}
	return defs
}

// ToolDef 用于传给 LLM 的工具定义（避免循环依赖 llm 包）。
type ToolDef struct {
	Function FunctionDef `json:"function"`
}

// FunctionDef 是工具的函数定义。
type FunctionDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}
