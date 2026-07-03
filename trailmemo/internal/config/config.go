package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	platformlogger "github.com/trailmemo/internal/platform/logger"
)

type Config struct {
	Server   ServerConfig          `mapstructure:"server"`
	Database DatabaseConfig        `mapstructure:"database"`
	Redis    RedisConfig           `mapstructure:"redis"`
	JWT      JWTConfig             `mapstructure:"jwt"`
	Wechat   WechatConfig          `mapstructure:"wechat"`
	Upload   UploadConfig          `mapstructure:"upload"`
	LLM      LLMConfig             `mapstructure:"llm"`
	Agent    AgentConfig           `mapstructure:"agent"`
	Map      MapConfig             `mapstructure:"map"`
	Log      platformlogger.Config `mapstructure:"log"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Name         string `mapstructure:"name"` // 数据库名称
	Charset      string `mapstructure:"charset"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	Loc          string `mapstructure:"loc"` // 时区设置为本地时区
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=%s",
		d.Username, d.Password, d.Host, d.Port, d.Name, d.Charset, d.Loc)
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type JWTConfig struct {
	Secret      string   `mapstructure:"secret"`
	ExpireHours int      `mapstructure:"expire_hours"`
	IgnorePaths []string `mapstructure:"ignore_paths"` // 忽略认证的路径
}

type WechatConfig struct {
	AppID     string `mapstructure:"appid"`
	AppSecret string `mapstructure:"appsecret"`
}

type UploadConfig struct {
	Dir     string `mapstructure:"dir"`
	MaxSize int    `mapstructure:"max_size"`
}

type LLMConfig struct {
	Provider string `mapstructure:"provider"`
	APIKey   string `mapstructure:"api_key"`
	BaseURL  string `mapstructure:"base_url"`
	Model    string `mapstructure:"model"`
}

type AgentConfig struct {
	Enabled        bool          `mapstructure:"enabled"`          // 是否启用智能体
	DefaultMode    string        `mapstructure:"default_mode"`     // 默认模式
	MaxSteps       int           `mapstructure:"max_steps"`        // 最大步骤数
	MaxToolCalls   int           `mapstructure:"max_tool_calls"`   // 最大工具调用次数
	MaxInputLength int           `mapstructure:"max_input_length"` // 输入最大长度（字符）
	RequestTimeout string        `mapstructure:"request_timeout"`  // 请求超时时间
	StreamTimeout  string        `mapstructure:"stream_timeout"`   // 流式响应超时时间
	LLM            LLMConfig     `mapstructure:"llm"`              // LLM配置
	Budget         AgentBudget   `mapstructure:"budget"`           // 预算配置
	Approval       AgentApproval `mapstructure:"approval"`         // 审批配置
	Cache          AgentCache    `mapstructure:"cache"`            // 缓存配置
}

type AgentBudget struct {
	MaxTokensPerRun      int `mapstructure:"max_tokens_per_run"`        // 单次运行最大 token 数
	MaxRunsPerUserPerDay int `mapstructure:"max_runs_per_user_per_day"` // 最大用户每天运行次数
}

type AgentApproval struct {
	RequireForPublicAction bool `mapstructure:"require_for_public_action"` // 是否需要审批公开操作
	RequireForUserWrite    bool `mapstructure:"require_for_user_write"`    // 是否需要审批用户写入操作
}

type AgentCache struct {
	RecommendationTTL string `mapstructure:"recommendation_ttl"` // 推荐缓存过期时间
	GuideTTL          string `mapstructure:"guide_ttl"`          // 攻略缓存过期时间
	SessionTTL        string `mapstructure:"session_ttl"`        // 会话缓存过期时间
}

type MapConfig struct {
	Provider string `mapstructure:"provider"`
	APIKey   string `mapstructure:"api_key"`
}

var globalConfig *Config

func Get() *Config {
	return globalConfig
}

func Load(configPath string) error {
	if configPath == "" {
		configPath = "./configs"
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	finalizeConfig(&cfg)

	globalConfig = &cfg
	return nil
}

func LoadWithFile(filename string) error {
	dir := filepath.Dir(filename)
	name := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))

	viper.SetConfigName(name)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(dir)

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	finalizeConfig(&cfg)

	globalConfig = &cfg
	return nil
}

func finalizeConfig(cfg *Config) {
	cfg.Log = platformlogger.WithDefaults(cfg.Log, cfg.Server.Mode)
	applySecretEnvOverrides(cfg)
	applyAgentDefaults(cfg)
}

func applySecretEnvOverrides(cfg *Config) {
	overrideString(&cfg.LLM.Provider, "TRAILMEMO_LLM_PROVIDER")
	overrideString(&cfg.LLM.APIKey, "TRAILMEMO_LLM_API_KEY")
	overrideString(&cfg.LLM.BaseURL, "TRAILMEMO_LLM_BASE_URL")
	overrideString(&cfg.LLM.Model, "TRAILMEMO_LLM_MODEL")

	overrideString(&cfg.Agent.LLM.Provider, "TRAILMEMO_AGENT_LLM_PROVIDER")
	overrideString(&cfg.Agent.LLM.APIKey, "TRAILMEMO_AGENT_LLM_API_KEY")
	overrideString(&cfg.Agent.LLM.BaseURL, "TRAILMEMO_AGENT_LLM_BASE_URL")
	overrideString(&cfg.Agent.LLM.Model, "TRAILMEMO_AGENT_LLM_MODEL")

	overrideString(&cfg.Map.APIKey, "TRAILMEMO_MAP_API_KEY")
	overrideString(&cfg.Wechat.AppID, "TRAILMEMO_WECHAT_APPID")
	overrideString(&cfg.Wechat.AppSecret, "TRAILMEMO_WECHAT_APPSECRET")
}

func applyAgentDefaults(cfg *Config) {
	if cfg.Agent.DefaultMode == "" {
		cfg.Agent.DefaultMode = "workflow_first"
	}
	if cfg.Agent.MaxSteps == 0 {
		cfg.Agent.MaxSteps = 6
	}
	if cfg.Agent.MaxToolCalls == 0 {
		cfg.Agent.MaxToolCalls = 10
	}
	if cfg.Agent.MaxInputLength == 0 {
		cfg.Agent.MaxInputLength = 5000
	}
	if cfg.Agent.RequestTimeout == "" {
		cfg.Agent.RequestTimeout = "20s"
	}
	if cfg.Agent.StreamTimeout == "" {
		cfg.Agent.StreamTimeout = "60s"
	}
	if cfg.Agent.LLM.Provider == "" {
		cfg.Agent.LLM.Provider = cfg.LLM.Provider
	}
	if cfg.Agent.LLM.APIKey == "" {
		cfg.Agent.LLM.APIKey = cfg.LLM.APIKey
	}
	if cfg.Agent.LLM.BaseURL == "" {
		cfg.Agent.LLM.BaseURL = cfg.LLM.BaseURL
	}
	if cfg.Agent.LLM.Model == "" {
		cfg.Agent.LLM.Model = cfg.LLM.Model
	}
	if cfg.Agent.Budget.MaxTokensPerRun == 0 {
		cfg.Agent.Budget.MaxTokensPerRun = 12000
	}
	if cfg.Agent.Budget.MaxRunsPerUserPerDay == 0 {
		cfg.Agent.Budget.MaxRunsPerUserPerDay = 100
	}
	if cfg.Agent.Cache.RecommendationTTL == "" {
		cfg.Agent.Cache.RecommendationTTL = "30m"
	}
	if cfg.Agent.Cache.GuideTTL == "" {
		cfg.Agent.Cache.GuideTTL = "12h"
	}
	if cfg.Agent.Cache.SessionTTL == "" {
		cfg.Agent.Cache.SessionTTL = "2h"
	}
}

func overrideString(target *string, envKey string) {
	if value := strings.TrimSpace(os.Getenv(envKey)); value != "" {
		*target = value
	}
}
