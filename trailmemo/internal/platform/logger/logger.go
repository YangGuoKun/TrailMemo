package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var baseLogger = zap.NewNop() // 日志器
var activeConfig Config

// Init 初始化全局 logger。启动阶段只做一次，后续通过 L()/FromContext() 获取。
func Init(cfg Config, baseFields ...zap.Field) error {
	activeConfig = cfg

	// 初始化日志级别
	level := zap.NewAtomicLevel() // 日志级别
	if err := level.UnmarshalText([]byte(strings.ToLower(cfg.Level))); err != nil {
		level.SetLevel(zap.InfoLevel)
	}

	encoderCfg := zap.NewProductionEncoderConfig()            // 编码器配置
	encoderCfg.TimeKey = "timestamp"                          // 时间键名
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder        // 时间格式
	encoderCfg.EncodeDuration = zapcore.MillisDurationEncoder // 持续时间格式

	var encoder zapcore.Encoder // 编码器
	if cfg.Format == "console" {
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder // 日志级别格式器
		encoder = zapcore.NewConsoleEncoder(encoderCfg)           // 控制台编码器
	} else {
		encoder = zapcore.NewJSONEncoder(encoderCfg) // JSON 编码器
	}

	writer := zapcore.AddSync(os.Stdout) // 输出流
	if cfg.Output == "file" && cfg.FilePath != "" {
		file, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return err
		}
		writer = zapcore.AddSync(file) // 文件流
	}

	core := zapcore.NewCore(encoder, writer, level)    // 核心日志器
	options := []zap.Option{zap.Fields(baseFields...)} // 日志选项
	if cfg.AddCaller {
		options = append(options, zap.AddCaller())
	}
	if parseLevel(cfg.StacktraceLevel) != nil {
		options = append(options, zap.AddStacktrace(*parseLevel(cfg.StacktraceLevel)))
	}

	baseLogger = zap.New(core, options...) // 日志器
	return nil
}

// L 返回全局基础 logger。没有请求上下文的启动、后台任务日志使用它。
func L() *zap.Logger {
	if baseLogger == nil {
		return zap.NewNop()
	}
	return baseLogger
}

// ConfigValue 返回当前日志配置。
func ConfigValue() Config {
	return activeConfig
}

// Sync 确保缓冲日志刷出。程序退出前调用即可。
func Sync() {
	if baseLogger != nil {
		_ = baseLogger.Sync() // 同步日志
	}
}

// parseLevel 解析日志级别。
func parseLevel(value string) *zapcore.Level {
	if value == "" {
		return nil
	}
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(strings.ToLower(value))); err != nil {
		return nil
	}
	return &level
}
