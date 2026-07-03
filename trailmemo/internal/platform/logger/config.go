package logger

// Config 是日志系统的配置入口，对应 configs/config.yaml 里的 log 节点。
type Config struct {
	Level           string        `mapstructure:"level"`
	Format          string        `mapstructure:"format"`
	Output          string        `mapstructure:"output"`
	FilePath        string        `mapstructure:"file_path"`
	AddCaller       bool          `mapstructure:"add_caller"`
	StacktraceLevel string        `mapstructure:"stacktrace_level"`
	Request         RequestConfig `mapstructure:"request"`
	Gorm            GormConfig    `mapstructure:"gorm"`
	Audit           AuditConfig   `mapstructure:"audit"`
}

type RequestConfig struct {
	EnableBody        bool     `mapstructure:"enable_body"`
	MaxBodySize       int      `mapstructure:"max_body_size"`
	SkipPaths         []string `mapstructure:"skip_paths"`
	SampleSuccessRate float64  `mapstructure:"sample_success_rate"`
}

// GormConfig 的 bool 字段使用指针，区分"YAML 未配置"（nil）与"显式设为 false"。
type GormConfig struct {
	Level                     string `mapstructure:"level"`
	SlowThresholdMilliseconds int    `mapstructure:"slow_threshold_ms"`
	IgnoreRecordNotFound      *bool  `mapstructure:"ignore_record_not_found"`
	ParameterizedQueries      *bool  `mapstructure:"parameterized_queries"`
}

type AuditConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

func DefaultConfig(mode string) Config {
	format := "json"
	level := "info"
	if mode == "debug" {
		format = "console"
		level = "debug"
	}

	return Config{
		Level:           level,
		Format:          format,
		Output:          "stdout",
		AddCaller:       true,
		StacktraceLevel: "error",
		Request: RequestConfig{
			EnableBody:        false,
			MaxBodySize:       2048,
			SkipPaths:         []string{"/health", "/swagger"},
			SampleSuccessRate: 1,
		},
		Gorm: GormConfig{
			Level:                     "warn",
			SlowThresholdMilliseconds: 500,
			IgnoreRecordNotFound:      boolPtr(true),
			ParameterizedQueries:      boolPtr(true),
		},
		Audit: AuditConfig{
			Enabled: true,
		},
	}
}

// WithDefaults 补齐 YAML 中未显式配置的字段。指针字段只在 nil 时才填默认值。
func WithDefaults(cfg Config, mode string) Config {
	defaults := DefaultConfig(mode)

	if cfg.Level == "" {
		cfg.Level = defaults.Level
	}
	if cfg.Format == "" {
		cfg.Format = defaults.Format
	}
	if cfg.Output == "" {
		cfg.Output = defaults.Output
	}
	if cfg.StacktraceLevel == "" {
		cfg.StacktraceLevel = defaults.StacktraceLevel
	}
	if cfg.Request.MaxBodySize == 0 {
		cfg.Request.MaxBodySize = defaults.Request.MaxBodySize
	}
	if cfg.Request.SampleSuccessRate == 0 {
		cfg.Request.SampleSuccessRate = defaults.Request.SampleSuccessRate
	}
	if len(cfg.Request.SkipPaths) == 0 {
		cfg.Request.SkipPaths = defaults.Request.SkipPaths
	}
	if cfg.Gorm.Level == "" {
		cfg.Gorm.Level = defaults.Gorm.Level
	}
	if cfg.Gorm.SlowThresholdMilliseconds == 0 {
		cfg.Gorm.SlowThresholdMilliseconds = defaults.Gorm.SlowThresholdMilliseconds
	}
	if cfg.Gorm.IgnoreRecordNotFound == nil {
		cfg.Gorm.IgnoreRecordNotFound = defaults.Gorm.IgnoreRecordNotFound
	}
	if cfg.Gorm.ParameterizedQueries == nil {
		cfg.Gorm.ParameterizedQueries = defaults.Gorm.ParameterizedQueries
	}

	return cfg
}

func boolPtr(b bool) *bool { return &b }
