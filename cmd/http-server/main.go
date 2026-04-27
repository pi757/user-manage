package main

import (
	"fmt"
	"log"
	"time"
	"user-management-system/config"
	"user-management-system/database"
	"user-management-system/discovery"
	"user-management-system/handlers"
	"user-management-system/middleware"
	"user-management-system/rpc"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载HTTP服务配置（仅包含需要的配置）
	cfg := config.LoadHTTPServiceConfig()

	// 初始化Redis
	if err := database.InitRedis(&cfg.Redis); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	// 创建服务发现
	serviceDisc := discovery.NewRedisServiceDiscovery(database.RedisClient, "")
	defer serviceDisc.Close()

	// 创建RPC客户端连接池（使用服务发现）
	rpcPool := rpc.NewClientPool(100, rpc.WithServiceDiscovery(serviceDisc, "rpc-tcp-server"))
	defer rpcPool.Close()

	// 创建处理器
	handler := handlers.NewHandler(rpcPool)
	uploadHandler := handlers.NewUploadHandler(rpcPool, cfg.FileUpload.UploadDir, cfg.FileUpload.MaxSize)

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin引擎
	router := gin.Default()

	// 速率限制中间件（每个IP每分钟最多60个请求）
	rateLimiter := middleware.NewRateLimiter(60, time.Minute)
	router.Use(rateLimiter.Middleware())

	// 跨域中间件
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// API路由
	api := router.Group("/api")
	{
		// 认证相关
		api.POST("/register", handler.Register)
		api.POST("/login", handler.Login)
		api.POST("/logout", handler.Logout)

		// 用户信息
		api.GET("/profile", handler.GetProfile)
		api.PUT("/profile", handler.UpdateProfile)

		// 文件上传
		api.POST("/upload/avatar", uploadHandler.UploadAvatar)
	}

	// 静态文件服务
	router.Static("/web", "./web")                      // 前端页面
	router.Static("/uploads", cfg.FileUpload.UploadDir) // 头像访问

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 启动HTTP服务器
	addr := cfg.HTTPServer.Port
	fmt.Printf("HTTP server started on %s\n", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
