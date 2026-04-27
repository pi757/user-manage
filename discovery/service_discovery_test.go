package discovery

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// setupTestRedis 创建测试用的Redis客户端
func setupTestRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6380",
		DB:   1, // 使用不同的DB避免污染生产数据
	})
}

func TestServiceDiscovery_RegisterAndDiscover(t *testing.T) {
	client := setupTestRedis()
	defer client.Close()

	sd := NewRedisServiceDiscovery(client, "test:service")
	ctx := context.Background()

	// 注册服务
	info := &ServiceInfo{
		ServiceName: "test-service",
		Address:     "localhost",
		Port:        9090,
		Metadata: map[string]string{
			"version": "1.0.0",
		},
	}

	ttl := 30 * time.Second
	err := sd.Register(ctx, info, ttl)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	// 发现服务
	services, err := sd.Discover(ctx, "test-service")
	if err != nil {
		t.Fatalf("Failed to discover service: %v", err)
	}

	if len(services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(services))
	}

	if services[0].Address != "localhost" || services[0].Port != 9090 {
		t.Errorf("Service info mismatch: %+v", services[0])
	}

	// 清理
	err = sd.Deregister(ctx, "test-service", "localhost:9090")
	if err != nil {
		t.Logf("Warning: Failed to deregister: %v", err)
	}
}

func TestServiceDiscovery_Heartbeat(t *testing.T) {
	client := setupTestRedis()
	defer client.Close()

	sd := NewRedisServiceDiscovery(client, "test:service")
	ctx := context.Background()

	// 注册服务
	info := &ServiceInfo{
		ServiceName: "test-service",
		Address:     "localhost",
		Port:        9091,
	}

	ttl := 5 * time.Second
	err := sd.Register(ctx, info, ttl)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	// 发送心跳
	address := "localhost:9091"
	err = sd.Heartbeat(ctx, "test-service", address, ttl)
	if err != nil {
		t.Fatalf("Failed to send heartbeat: %v", err)
	}

	// 验证服务仍然健康
	services, err := sd.Discover(ctx, "test-service")
	if err != nil {
		t.Fatalf("Failed to discover service: %v", err)
	}

	if len(services) != 1 {
		t.Errorf("Expected 1 healthy service, got %d", len(services))
	}

	// 清理
	err = sd.Deregister(ctx, "test-service", address)
	if err != nil {
		t.Logf("Warning: Failed to deregister: %v", err)
	}
}

func TestServiceDiscovery_MultipleInstances(t *testing.T) {
	client := setupTestRedis()
	defer client.Close()

	sd := NewRedisServiceDiscovery(client, "test:service")
	ctx := context.Background()

	// 注册多个服务实例
	instances := []ServiceInfo{
		{ServiceName: "test-service", Address: "localhost", Port: 9090},
		{ServiceName: "test-service", Address: "localhost", Port: 9091},
		{ServiceName: "test-service", Address: "localhost", Port: 9092},
	}

	ttl := 30 * time.Second
	for _, info := range instances {
		err := sd.Register(ctx, &info, ttl)
		if err != nil {
			t.Fatalf("Failed to register service %d: %v", info.Port, err)
		}
	}

	// 发现所有服务
	services, err := sd.Discover(ctx, "test-service")
	if err != nil {
		t.Fatalf("Failed to discover services: %v", err)
	}

	if len(services) != 3 {
		t.Errorf("Expected 3 services, got %d", len(services))
	}

	// 发现一个服务（负载均衡）
	service, err := sd.DiscoverOne(ctx, "test-service")
	if err != nil {
		t.Fatalf("Failed to discover one service: %v", err)
	}

	if service == nil {
		t.Error("Expected non-nil service")
	}

	// 清理
	for _, info := range instances {
		address := fmt.Sprintf("%s:%d", info.Address, info.Port)
		err = sd.Deregister(ctx, "test-service", address)
		if err != nil {
			t.Logf("Warning: Failed to deregister %s: %v", address, err)
		}
	}
}

func TestServiceDiscovery_ServiceExpiration(t *testing.T) {
	client := setupTestRedis()
	defer client.Close()

	sd := NewRedisServiceDiscovery(client, "test:service")
	ctx := context.Background()

	// 注册服务，TTL很短
	info := &ServiceInfo{
		ServiceName: "test-service",
		Address:     "localhost",
		Port:        9093,
	}

	ttl := 2 * time.Second
	err := sd.Register(ctx, info, ttl)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	// 等待服务过期
	time.Sleep(3 * time.Second)

	// 尝试发现服务，应该失败
	_, err = sd.Discover(ctx, "test-service")
	if err == nil {
		t.Error("Expected error for expired service, got nil")
	}
}
