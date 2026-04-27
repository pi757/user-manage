package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	"user-management-system/auth"
	"user-management-system/config"
	"user-management-system/database"
	"user-management-system/discovery"
	"user-management-system/rpc"
	"user-management-system/service"
)

func main() {
	// 加载TCP服务配置（仅包含需要的配置）
	cfg := config.LoadTCPServiceConfig()

	// 初始化数据库
	if err := database.InitMySQL(&cfg.MySQL); err != nil {
		log.Fatalf("Failed to initialize MySQL: %v", err)
	}

	// 初始化Redis
	if err := database.InitRedis(&cfg.Redis); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	// 创建服务发现
	serviceDisc := discovery.NewRedisServiceDiscovery(database.RedisClient, "")
	defer serviceDisc.Close()

	// 解析TCP服务器地址
	tcpAddr := cfg.TCPServer.Port
	if tcpAddr[0] == ':' {
		tcpAddr = "0.0.0.0" + tcpAddr
	}
	host, portStr, err := net.SplitHostPort(tcpAddr)
	if err != nil {
		log.Fatalf("Invalid TCP server address: %v", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid TCP server port: %v", err)
	}

	// 注册服务到服务发现
	serviceInfo := &discovery.ServiceInfo{
		ServiceName: "rpc-tcp-server",
		Address:     host,
		Port:        port,
		Metadata: map[string]string{
			"version": "1.0.0",
		},
	}

	ctx := context.Background()
	ttl := 30 * time.Second // TTL 30秒

	if err := serviceDisc.Register(ctx, serviceInfo, ttl); err != nil {
		log.Printf("Warning: Failed to register service: %v", err)
	} else {
		fmt.Printf("Service registered: %s:%d\n", host, port)
	}

	// 启动心跳协程
	go func() {
		ticker := time.NewTicker(10 * time.Second) // 每10秒发送一次心跳
		defer ticker.Stop()

		address := fmt.Sprintf("%s:%d", host, port)
		for range ticker.C {
			if err := serviceDisc.Heartbeat(ctx, "rpc-tcp-server", address, ttl); err != nil {
				log.Printf("Warning: Heartbeat failed: %v", err)
			}
		}
	}()

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

	// 注销服务
	address := fmt.Sprintf("%s:%d", host, port)
	if err := serviceDisc.Deregister(ctx, "rpc-tcp-server", address); err != nil {
		log.Printf("Warning: Failed to deregister service: %v", err)
	} else {
		fmt.Println("Service deregistered")
	}

	err = rpcServer.Stop()
	if err != nil {
		return
	}
}
