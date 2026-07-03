package logger

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	gormlogger "gorm.io/gorm/logger"
)

// GormLogger 把 GORM 的慢 SQL 和错误 SQL 接入 zap。
type GormLogger struct {
	level                gormlogger.LogLevel // 日志级别
	slowThreshold        time.Duration       // 慢 SQL 阈值
	ignoreRecordNotFound bool                // 是否忽略 RecordNotFound 错误
}

// NewGormLogger 创建 GormLogger。
func NewGormLogger(cfg GormConfig) gormlogger.Interface {
	ignoreNotFound := false
	if cfg.IgnoreRecordNotFound != nil {
		ignoreNotFound = *cfg.IgnoreRecordNotFound
	}
	return &GormLogger{
		level:                parseGormLevel(cfg.Level),
		slowThreshold:        time.Duration(cfg.SlowThresholdMilliseconds) * time.Millisecond,
		ignoreRecordNotFound: ignoreNotFound,
	}
}

// LogMode 设置日志级别。
func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	clone := *l
	clone.level = level
	return &clone
}

// Info 记录信息日志。
func (l *GormLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	if l.level >= gormlogger.Info {
		FromContext(ctx).Sugar().Infof(msg, args...)
	}
}

// Warn 记录警告日志。
func (l *GormLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	if l.level >= gormlogger.Warn {
		FromContext(ctx).Sugar().Warnf(msg, args...)
	}
}

// Error 记录错误日志。
func (l *GormLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	if l.level >= gormlogger.Error {
		FromContext(ctx).Sugar().Errorf(msg, args...)
	}
}

// Trace 记录 GORM 查询日志。
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level == gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	fields := []zap.Field{
		zap.String("sql", sql),
		zap.Int64("rows", rows),
		zap.Int64("latency_ms", elapsed.Milliseconds()),
	}

	log := FromContext(ctx)
	if err != nil && !(l.ignoreRecordNotFound && errors.Is(err, gormlogger.ErrRecordNotFound)) {
		log.Error("gorm_error", append(fields,
			zap.String("event", "gorm_error"),
			zap.String("error_kind", "db"),
			zap.Error(err),
		)...)
		return
	}

	if l.slowThreshold > 0 && elapsed > l.slowThreshold && l.level >= gormlogger.Warn {
		log.Warn("gorm_slow_query", append(fields,
			zap.String("event", "gorm_slow_query"),
			zap.Duration("slow_threshold", l.slowThreshold),
		)...)
		return
	}

	if l.level >= gormlogger.Info {
		log.Debug("gorm_query", append(fields, zap.String("event", "gorm_query"))...)
	}
}

// parseGormLevel 解析 GORM 日志级别。
func parseGormLevel(value string) gormlogger.LogLevel {
	switch value {
	case "silent":
		return gormlogger.Silent
	case "error":
		return gormlogger.Error
	case "info", "debug":
		return gormlogger.Info
	default:
		return gormlogger.Warn
	}
}
