package runtime

// Intent identifies the business purpose of an agent run.
type Intent string // 意图

const (
	IntentUnknown    Intent = "unknown"
	IntentChat       Intent = "chat"
	IntentRecommend  Intent = "recommend"
	IntentRouteDraft Intent = "route_draft" // 路由草稿
	IntentRouteRemix Intent = "route_remix" // 路由 Remix
	IntentTravelNote Intent = "travel_note" // 旅行笔记
	IntentModeration Intent = "moderation"  // 审批
)

// RunStatus is the persisted lifecycle state of an agent run.
type RunStatus string // 运行状态

const (
	RunStatusCreated         RunStatus = "created"          // 创建
	RunStatusContextLoading  RunStatus = "context_loading"  // 上下文加载中
	RunStatusIntentResolved  RunStatus = "intent_resolved"  // 意图解析完成
	RunStatusRunning         RunStatus = "running"          // 运行中
	RunStatusToolCalling     RunStatus = "tool_calling"     // 工具调用中
	RunStatusValidating      RunStatus = "validating"       // 验证中
	RunStatusApprovalPending RunStatus = "approval_pending" // 审批待处理
	RunStatusCompleted       RunStatus = "completed"        // 完成
	RunStatusFailed          RunStatus = "failed"           // 失败
	RunStatusCancelled       RunStatus = "cancelled"        // 已取消
)

// ExecutionMode describes how a request should be handled.
type ExecutionMode string // 执行模式

const (
	ExecutionModeOneShot       ExecutionMode = "one_shot"       // 单次运行
	ExecutionModeWorkflow      ExecutionMode = "workflow"       // 工作流
	ExecutionModeAgentLoop     ExecutionMode = "agent_loop"     // 代理循环
	ExecutionModeWorkflowFirst ExecutionMode = "workflow_first" // 工作流首次运行
)

// StepType describes one recorded step in an agent run.
type StepType string // 步骤类型

const (
	StepTypeLLM        StepType = "llm"        // LLM 调用
	StepTypeTool       StepType = "tool"       // 工具调用
	StepTypeValidation StepType = "validation" // 验证
	StepTypeApproval   StepType = "approval"   // 审批
	StepTypeFallback   StepType = "fallback"   // 回退
)
