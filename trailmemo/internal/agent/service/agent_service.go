package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/trailmemo/internal/agent/dto"
	"github.com/trailmemo/internal/agent/guardrail"
	"github.com/trailmemo/internal/agent/llm"
	"github.com/trailmemo/internal/agent/memory"
	"github.com/trailmemo/internal/agent/prompt"
	agentruntime "github.com/trailmemo/internal/agent/runtime"
	"github.com/trailmemo/internal/agent/tools"
	"github.com/trailmemo/internal/agent/workflow"
	"github.com/trailmemo/internal/config"
	"github.com/trailmemo/internal/platform/logger"
	agentservice2 "github.com/trailmemo/internal/service"
	"go.uber.org/zap"
)

// AgentService 是 Agent 模块的应用服务层，连接 HTTP Handler 与 Agent 内部组件。
// 负责初始化 LLM 客户端、工具注册中心、记忆系统和 Prompt 管理器。
// 具体业务方法拆分到各独立 service 文件。
type AgentService struct {
	cfg          config.AgentConfig
	llmClient    *llm.Client
	promptMgr    *prompt.Manager
	validator    *guardrail.Validator
	toolReg      *tools.Registry
	runStore     *memory.RunStore
	artStore     *memory.ArtifactStore
	sessMem      *memory.SessionMemory
	guardrail    *guardrail.Service
	prefStore    *memory.PreferenceStore
	sessionStore *memory.SessionStore
}

// NewAgentService 创建 Agent 服务实例，执行首次数据库迁移并注册首批只读工具。
func NewAgentService() *AgentService {
	cfg := config.Get().Agent
	svc := &AgentService{
		cfg:          cfg,
		llmClient:    llm.NewClient(cfg.LLM),
		promptMgr:    prompt.NewManager(prompt.MapLoader(defaultPrompts())),
		validator:    guardrail.NewValidator(1),
		toolReg:      tools.NewRegistry(),
		runStore:     memory.NewRunStore(),
		artStore:     memory.NewArtifactStore(),
		sessMem:      memory.NewSessionMemory(parseDuration(cfg.Cache.SessionTTL, 2*time.Hour)),
		guardrail:    guardrail.NewService(cfg),
		prefStore:    memory.NewPreferenceStore(),
		sessionStore: memory.NewSessionStore(),
	}
	if err := svc.migrate(); err != nil {
		logger.L().Warn("agent_migration_skipped", zap.Error(err))
	}
	svc.registerTools()
	logger.L().Info("agent_service_initialized",
		zap.Bool("enabled", cfg.Enabled),
		zap.String("model", cfg.LLM.Model),
		zap.Int("tools_registered", len(svc.toolReg.GetAllDescriptors())),
	)
	return svc
}

func (s *AgentService) migrate() error {
	db := config.GetDB()
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}
	return db.AutoMigrate(&memory.AgentRun{}, &memory.AgentStep{}, &memory.AgentArtifact{}, &memory.AgentUserPreference{}, &memory.AgentSession{})
}

func (s *AgentService) registerTools() {
	s.toolReg.Register(tools.NewRouteTool(agentservice2.NewRouteService()))
	s.toolReg.Register(tools.NewCheckinTool(agentservice2.NewCheckinService()))
	s.toolReg.Register(tools.NewCommunityTool(agentservice2.NewPostService()))
	mapCfg := config.Get().Map
	if strings.EqualFold(strings.TrimSpace(mapCfg.Provider), "amap") && strings.TrimSpace(mapCfg.APIKey) != "" {
		s.toolReg.Register(tools.NewMapPOIToolWithSearcher("amap", tools.NewAmapPOISearcher(mapCfg.APIKey)))
		logger.L().Info("agent_map_poi_tool_registered", zap.String("source", "amap"))
		return
	}
	s.toolReg.Register(tools.NewMapPOITool())
	logger.L().Info("agent_map_poi_tool_registered", zap.String("source", "local_seed"))
}

func (s *AgentService) Health() dto.HealthResponse {
	return dto.HealthResponse{Status: "ok", Enabled: s.cfg.Enabled, Stage: "phase3",
		LLMConfigured: s.cfg.LLM.APIKey != "", DefaultMode: s.cfg.DefaultMode,
		RequestTimeout: s.cfg.RequestTimeout, StreamTimeout: s.cfg.StreamTimeout}
}

func (s *AgentService) Capabilities() dto.CapabilityResponse {
	return dto.CapabilityResponse{Enabled: s.cfg.Enabled, DefaultMode: s.cfg.DefaultMode,
		MaxSteps: s.cfg.MaxSteps, MaxToolCalls: s.cfg.MaxToolCalls, Stage: "phase3",
		Intents: []string{string(agentruntime.IntentChat), string(agentruntime.IntentRecommend),
			string(agentruntime.IntentRouteDraft), string(agentruntime.IntentRouteRemix),
			string(agentruntime.IntentTravelNote), string(agentruntime.IntentModeration)},
		Tools: s.toolReg.GetAllDescriptors()}
}

func (s *AgentService) Chat(ctx context.Context, userMessage string) (*llm.ChatResponse, error) {
	if !s.cfg.Enabled {
		return nil, fmt.Errorf("agent 未启用")
	}
	if s.cfg.LLM.APIKey == "" {
		return nil, fmt.Errorf("LLM API key 未配置")
	}
	return s.llmClient.Chat(ctx, llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "system", Content: "你是迹忆旅图的AI旅行助手。"},
			{Role: "user", Content: userMessage}},
		MaxTokens: 1000})
}

// buildWorkflowContext 构造标准 WorkflowContext，加载用户偏好快照。
func (s *AgentService) buildWorkflowContext(ctx context.Context, userID uint64, intent agentruntime.Intent, mode agentruntime.ExecutionMode) *workflow.WorkflowContext {
	snapshot := s.prefStore.GetSnapshot(ctx, userID)
	return &workflow.WorkflowContext{
		Ctx: ctx, RunID: uuid.NewString(), UserID: userID, Intent: intent, Mode: mode,
		LLMClient: s.llmClient, PromptMgr: s.promptMgr, ToolRegistry: s.toolReg,
		RunStore: s.runStore, ArtifactStore: s.artStore, SessionMem: s.sessMem,
		Guardrail: s.guardrail, Logger: logger.FromContext(ctx),
		Preferences: s.convertSnapshot(snapshot),
	}
}

func (s *AgentService) convertSnapshot(sn *memory.PreferenceSnapshot) *workflow.UserPrefs {
	return &workflow.UserPrefs{
		TravelStyles: nil, BudgetLevel: sn.BudgetLevel, Interests: sn.Interests, Pace: sn.Pace,
	}
}

// ── 共享转换工具 ──────────────────────────────────

func convertRouteDraftToDTO(a *workflow.RouteDraftArtifact) *dto.RouteDraftData {
	cps := make([]dto.CheckpointDraftData, 0, len(a.Checkpoints))
	for _, cp := range a.Checkpoints {
		cps = append(cps, dto.CheckpointDraftData{Name: cp.Name, Description: cp.Description,
			City: cp.City, Address: cp.Address, Latitude: cp.Latitude, Longitude: cp.Longitude,
			Sequence: cp.Sequence, ArriveTime: cp.ArriveTime, StayDuration: cp.StayDuration})
	}
	return &dto.RouteDraftData{Title: a.Title, Summary: a.Summary, StartCity: a.StartCity,
		EndCity: a.EndCity, EstimatedBudget: a.EstimatedBudget, EstimatedHours: a.EstimatedHours, Checkpoints: cps}
}

func convertRemixToDTO(a *workflow.RemixArtifact) *dto.RouteDraftData {
	cps := make([]dto.CheckpointDraftData, 0, len(a.Checkpoints))
	for _, cp := range a.Checkpoints {
		cps = append(cps, dto.CheckpointDraftData{Name: cp.Name, Description: cp.Description,
			City: cp.City, Address: cp.Address, Latitude: cp.Latitude, Longitude: cp.Longitude,
			Sequence: cp.Sequence, ArriveTime: cp.ArriveTime, StayDuration: cp.StayDuration})
	}
	return &dto.RouteDraftData{Title: a.Title, Summary: a.Summary, StartCity: a.StartCity,
		EndCity: a.EndCity, EstimatedBudget: a.EstimatedBudget, EstimatedHours: a.EstimatedHours, Checkpoints: cps}
}

func parseUint64(s string) uint64 { var id uint64; fmt.Sscanf(s, "%d", &id); return id }

func parseDuration(s string, defaultVal time.Duration) time.Duration {
	if s == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return defaultVal
	}
	return d
}

func defaultPrompts() map[string]string {
	return map[string]string{
		"route_draft": `你是旅行路线规划师。根据用户需求生成路线草稿。需求：{{user_query}}。偏好：{{user_preferences}}。输出JSON：{title,summary,start_city,end_city,estimated_budget,checkpoints:[{name,city,sequence,arrive_time,stay_duration,description}]}。只输出JSON。`,
		"recommend":   `你是旅行推荐顾问。根据用户需求输出JSON数组，每条含title/description/city/estimated_budget/suggested_days/tags。需求：{{user_query}}。偏好：{{user_preferences}}。只输出JSON数组。`,
	}
}
