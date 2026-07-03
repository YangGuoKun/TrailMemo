package middleware

import (
	"github.com/gin-gonic/gin"
	platformlogger "github.com/trailmemo/internal/platform/logger"
	"go.uber.org/zap"
)

// InitLogger 保留旧入口，方便未迁移代码继续编译。
func InitLogger(mode string) error {
	return platformlogger.Init(platformlogger.DefaultConfig(mode),
		zap.String("service", "trailmemo-api"),
		zap.String("env", mode),
	)
}

// GetLogger 获取日志记录器
func GetLogger() *zap.Logger {
	return platformlogger.L()
}

// Logger 保留旧中间件名称，内部使用新版 AccessLog。
func Logger() gin.HandlerFunc {
	return platformlogger.AccessLog()
}

// SyncLogger 同步日志记录器，确保所有日志写入文件
func SyncLogger() {
	platformlogger.Sync()
}
