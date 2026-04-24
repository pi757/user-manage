package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"user-management-system/auth"
	"user-management-system/config"
	"user-management-system/database"
	"user-management-system/rpc"
	"user-management-system/service"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化数据库
	if err := database.InitMySQL(&cfg.MySQL); err != nil {
		log.Fatalf("Failed to initialize MySQL: %v", err)
	}

	// 初始化Redis
	if err := database.InitRedis(&cfg.Redis); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	// 创建认证服务
	authService := auth.NewAuthService(&cfg.Session)

	// 创建用户服务
	userService := service.NewUserService(authService)

	// 创建RPC服务器
	rpcServer := rpc.NewServer()

	// 注册RPC方法
	rpcServer.Register("user.register", func(params map[string]interface{}) (interface{}, error) {
		return userService.Register(params)
	})

	rpcServer.Register("user.login", func(params map[string]interface{}) (interface{}, error) {
		return userService.Login(params)
	})

	rpcServer.Register("user.getProfile", func(params map[string]interface{}) (interface{}, error) {
		return userService.GetProfile(params)
	})

	rpcServer.Register("user.updateProfile", func(params map[string]interface{}) (interface{}, error) {
		return userService.UpdateProfile(params)
	})

	rpcServer.Register("user.logout", func(params map[string]interface{}) (interface{}, error) {
		return userService.Logout(params)
	})

	rpcServer.Register("user.getUserByID", func(params map[string]interface{}) (interface{}, error) {
		return userService.GetUserByID(params)
	})

	rpcServer.Register("user.batchGetUsers", func(params map[string]interface{}) (interface{}, error) {
		return userService.BatchGetUsers(params)
	})

	rpcServer.Register("user.validateToken", func(params map[string]interface{}) (interface{}, error) {
		return userService.ValidateToken(params)
	})

	// 启动RPC服务器
	go func() {
		if err := rpcServer.Start(cfg.TCPServer.Port); err != nil {
			log.Fatalf("Failed to start RPC server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")
	rpcServer.Stop()
}
