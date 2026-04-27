package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ServiceInfo 服务信息
type ServiceInfo struct {
	ServiceName string            `json:"service_name"`
	Address     string            `json:"address"`
	Port        int               `json:"port"`
	Health      string            `json:"health"`             // healthy, unhealthy
	RegisterAt  time.Time         `json:"register_at"`        // 注册时间
	LastHeartAt time.Time         `json:"last_heart_at"`      // 最后心跳时间
	Metadata    map[string]string `json:"metadata,omitempty"` // 元数据
}

// ServiceDiscovery 服务发现接口
type ServiceDiscovery interface {
	// Register 注册服务
	Register(ctx context.Context, info *ServiceInfo, ttl time.Duration) error
	// Deregister 注销服务
	Deregister(ctx context.Context, serviceName, address string) error
	// Heartbeat 发送心跳
	Heartbeat(ctx context.Context, serviceName, address string, ttl time.Duration) error
	// Discover 发现服务（返回所有健康的服务实例）
	Discover(ctx context.Context, serviceName string) ([]*ServiceInfo, error)
	// DiscoverOne 发现一个服务实例（负载均衡）
	DiscoverOne(ctx context.Context, serviceName string) (*ServiceInfo, error)
	// Close 关闭连接
	Close() error
}

// RedisServiceDiscovery 基于Redis的服务发现实现
type RedisServiceDiscovery struct {
	client *redis.Client
	prefix string // Redis key前缀
}

// NewRedisServiceDiscovery 创建基于Redis的服务发现
func NewRedisServiceDiscovery(redisClient *redis.Client, prefix string) *RedisServiceDiscovery {
	if prefix == "" {
		prefix = "service:discovery"
	}
	return &RedisServiceDiscovery{
		client: redisClient,
		prefix: prefix,
	}
}

// Register 注册服务
func (d *RedisServiceDiscovery) Register(ctx context.Context, info *ServiceInfo, ttl time.Duration) error {
	info.Health = "healthy"
	info.RegisterAt = time.Now()
	info.LastHeartAt = time.Now()

	key := fmt.Sprintf("%s:%s:%s:%d", d.prefix, info.ServiceName, info.Address, info.Port)

	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal service info: %w", err)
	}

	// 设置服务信息和TTL
	if err := d.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	// 将服务地址添加到服务列表
	listKey := fmt.Sprintf("%s:%s:list", d.prefix, info.ServiceName)
	member := fmt.Sprintf("%s:%d", info.Address, info.Port)
	if err := d.client.SAdd(ctx, listKey, member).Err(); err != nil {
		return fmt.Errorf("failed to add service to list: %w", err)
	}

	return nil
}

// Deregister 注销服务
func (d *RedisServiceDiscovery) Deregister(ctx context.Context, serviceName, address string) error {
	// 从列表中移除
	listKey := fmt.Sprintf("%s:%s:list", d.prefix, serviceName)
	if err := d.client.SRem(ctx, listKey, address).Err(); err != nil {
		return fmt.Errorf("failed to remove service from list: %w", err)
	}

	// 删除服务详细信息
	key := fmt.Sprintf("%s:%s:%s", d.prefix, serviceName, address)
	if err := d.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete service info: %w", err)
	}

	return nil
}

// Heartbeat 发送心跳
func (d *RedisServiceDiscovery) Heartbeat(ctx context.Context, serviceName, address string, ttl time.Duration) error {
	key := fmt.Sprintf("%s:%s:%s", d.prefix, serviceName, address)

	// 检查服务是否存在
	exists, err := d.client.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to check service existence: %w", err)
	}

	if exists == 0 {
		return fmt.Errorf("service not found: %s:%s", serviceName, address)
	}

	// 更新最后心跳时间和TTL
	var info ServiceInfo
	data, err := d.client.Get(ctx, key).Bytes()
	if err != nil {
		return fmt.Errorf("failed to get service info: %w", err)
	}

	if err := json.Unmarshal(data, &info); err != nil {
		return fmt.Errorf("failed to unmarshal service info: %w", err)
	}

	info.LastHeartAt = time.Now()
	info.Health = "healthy"

	updatedData, err := json.Marshal(&info)
	if err != nil {
		return fmt.Errorf("failed to marshal updated service info: %w", err)
	}

	if err := d.client.Set(ctx, key, updatedData, ttl).Err(); err != nil {
		return fmt.Errorf("failed to update heartbeat: %w", err)
	}

	return nil
}

// Discover 发现服务（返回所有健康的服务实例）
func (d *RedisServiceDiscovery) Discover(ctx context.Context, serviceName string) ([]*ServiceInfo, error) {
	listKey := fmt.Sprintf("%s:%s:list", d.prefix, serviceName)

	// 获取所有服务地址
	members, err := d.client.SMembers(ctx, listKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get service list: %w", err)
	}

	if len(members) == 0 {
		return nil, fmt.Errorf("no services found for: %s", serviceName)
	}

	var healthyServices []*ServiceInfo

	for _, member := range members {
		key := fmt.Sprintf("%s:%s:%s", d.prefix, serviceName, member)

		data, err := d.client.Get(ctx, key).Bytes()
		if err != nil {
			// 跳过不存在的服务（可能已过期）
			continue
		}

		var info ServiceInfo
		if err := json.Unmarshal(data, &info); err != nil {
			continue
		}

		// 只返回健康的服务
		if info.Health == "healthy" {
			healthyServices = append(healthyServices, &info)
		}
	}

	if len(healthyServices) == 0 {
		return nil, fmt.Errorf("no healthy services found for: %s", serviceName)
	}

	return healthyServices, nil
}

// DiscoverOne 发现一个服务实例（简单轮询负载均衡）
func (d *RedisServiceDiscovery) DiscoverOne(ctx context.Context, serviceName string) (*ServiceInfo, error) {
	services, err := d.Discover(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	// 简单轮询：取第一个
	return services[0], nil
}

// Close 关闭连接
func (d *RedisServiceDiscovery) Close() error {
	return d.client.Close()
}
