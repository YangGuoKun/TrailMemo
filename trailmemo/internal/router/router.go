package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/trailmemo/docs"
	agenthandler "github.com/trailmemo/internal/agent/handler"
	"github.com/trailmemo/internal/config"
	"github.com/trailmemo/internal/handler"
	"github.com/trailmemo/internal/middleware"
	platformlogger "github.com/trailmemo/internal/platform/logger"
	"github.com/trailmemo/pkg/response"
)

func NewRouter() *gin.Engine {
	r := gin.New()

	r.Use(platformlogger.RequestID()) // 请求 ID 中间件
	r.Use(platformlogger.Recovery())  // 恢复中间件
	r.Use(platformlogger.AccessLog()) // 访问日志中间件
	r.Use(middleware.CORS())          // CORS 中间件

	cfg := config.Get()
	r.Static("/uploads", cfg.Upload.Dir)

	r.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{
			"status":  "ok",
			"service": "trailmemo",
		})
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", func(c *gin.Context) {
			response.Success(c, "pong")
		})

		userHandler := handler.NewUserHandler()
		userHandler.RegisterRoutes(v1)

		uploadHandler := handler.NewUploadHandler()
		uploadHandler.RegisterRoutes(v1)

		routeHandler := handler.NewRouteHandler()
		routeHandler.RegisterRoutes(v1)

		checkinHandler := handler.NewCheckinHandler()
		checkinHandler.RegisterRoutes(v1)

		postHandler := handler.NewPostHandler()
		postHandler.RegisterRoutes(v1)

		commentHandler := handler.NewCommentHandler()
		commentHandler.RegisterRoutes(v1)

		likeHandler := handler.NewLikeHandler()
		likeHandler.RegisterRoutes(v1)

		favoriteHandler := handler.NewFavoriteHandler()
		favoriteHandler.RegisterRoutes(v1)

		agentHandler := agenthandler.NewHandler()
		agentHandler.RegisterRoutes(v1)
	}

	return r
}
