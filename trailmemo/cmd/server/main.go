package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	_ "github.com/trailmemo/docs" // 导入 Swagger 文档
	"github.com/trailmemo/internal/config"
	"github.com/trailmemo/internal/middleware"
	"github.com/trailmemo/internal/model"
	platformlogger "github.com/trailmemo/internal/platform/logger"
	"github.com/trailmemo/internal/router"
	"go.uber.org/zap"
)

// @title TrailMemo API
// @version 1.0.0
// @description 迹忆旅图寻迹旅游后端 API 文档
// @termsOfService https://example.com/terms
// @contact.name API Support
// @contact.email support@example.com
// @license.name Apache 2.0
// @license.url https://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8087
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// 加载配置
	configPath := flag.String("config", "./configs", "config file path")
	flag.Parse()

	if err := config.Load(*configPath); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	cfg := config.Get() // 获取配置

	// 初始化日志。配置加载前仍使用标准库 log，初始化后统一改用 zap。
	if err := platformlogger.Init(cfg.Log,
		zap.String("service", "trailmemo-api"),
		zap.String("env", cfg.Server.Mode),
	); err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}
	defer platformlogger.Sync()

	// 初始化数据库
	if err := config.InitDB(&cfg.Database); err != nil {
		platformlogger.L().Fatal("database_init_failed",
			zap.String("event", "database_init_failed"),
			zap.Error(err),
		)
	}
	defer config.CloseDB()

	// 初始化Redis
	if err := config.InitRedis(&cfg.Redis); err != nil {
		platformlogger.L().Warn("redis_unavailable",
			zap.String("event", "redis_unavailable"),
			zap.String("message", "cache will be disabled"),
			zap.Error(err),
		)
	} else {
		defer config.CloseRedis()
	}

	// 自动迁移数据库
	if err := model.AutoMigrate(); err != nil {
		platformlogger.L().Fatal("migration_failed",
			zap.String("event", "migration_failed"),
			zap.Error(err),
		)
	}

	// 初始化JWT
	middleware.InitJWT()

	// 设置Gin模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode) // 设置为发布模式
	}

	// 创建路由
	r := router.NewRouter()

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	platformlogger.L().Info("server_started",
		zap.String("event", "server_started"),
		zap.String("addr", addr),
		zap.String("mode", cfg.Server.Mode),
	)

	quit := make(chan os.Signal, 1) // 用于接收信号
	// 监听SIGINT和SIGTERM信号, 用于优雅关闭服务器
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// 启动服务器
	go func() {
		if err := r.Run(addr); err != nil {
			platformlogger.L().Fatal("server_start_failed",
				zap.String("event", "server_start_failed"),
				zap.Error(err),
			)
		}
	}()

	<-quit // 等待信号
	platformlogger.L().Info("server_stopped",
		zap.String("event", "server_stopped"),
	)
}
