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

// RedisServiceDiscovery 基于Redis Sorted Set的服务发现实现
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

// Register 注册服务（使用Sorted Set）
func (d *RedisServiceDiscovery) Register(ctx context.Context, info *ServiceInfo, ttl time.Duration) error {
	info.Health = "healthy"
	info.RegisterAt = time.Now()
	info.LastHeartAt = time.Now()

	member := fmt.Sprintf("%s:%d", info.Address, info.Port)
	key := fmt.Sprintf("%s:%s:%s", d.prefix, info.ServiceName, member)

	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal service info: %w", err)
	}
	// 使用Pipeline保证原子性
	pipe := d.client.Pipeline()

	// 1. 设置服务详情和TTL
	pipe.Set(ctx, key, data, ttl)

	// 2. 添加到Sorted Set，score为当前时间戳（用于清理过期服务）
	zsetKey := fmt.Sprintf("%s:%s:zset", d.prefix, info.ServiceName)
	score := float64(time.Now().Unix())
	pipe.ZAdd(ctx, zsetKey, redis.Z{
		Score:  score,
		Member: member,
	})

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	return nil
}

// Deregister 注销服务
func (d *RedisServiceDiscovery) Deregister(ctx context.Context, serviceName, address string) error {
	// 使用Pipeline保证原子性
	pipe := d.client.Pipeline()

	// 1. 从Sorted Set中移除
	zsetKey := fmt.Sprintf("%s:%s:zset", d.prefix, serviceName)
	pipe.ZRem(ctx, zsetKey, address)

	// 2. 删除服务详细信息
	key := fmt.Sprintf("%s:%s:%s", d.prefix, serviceName, address)
	pipe.Del(ctx, key)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
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

	// 使用Pipeline更新
	pipe := d.client.Pipeline()

	// 1. 更新服务详情和TTL
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

	pipe.Set(ctx, key, updatedData, ttl)

	// 2. 更新Sorted Set的score为当前时间戳
	zsetKey := fmt.Sprintf("%s:%s:zset", d.prefix, serviceName)
	score := float64(time.Now().Unix())
	pipe.ZAdd(ctx, zsetKey, redis.Z{
		Score:  score,
		Member: address,
	})

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update heartbeat: %w", err)
	}

	return nil
}

// cleanupExpiredServices 清理过期的服务（score超过ttl的）
func (d *RedisServiceDiscovery) cleanupExpiredServices(ctx context.Context, serviceName string, ttl time.Duration) error {
	zsetKey := fmt.Sprintf("%s:%s:zset", d.prefix, serviceName)

	// 计算过期时间戳
	expiredAt := time.Now().Add(-ttl).Unix()

	// 移除所有score小于过期时间戳的成员
	removed, err := d.client.ZRemRangeByScore(ctx, zsetKey, "0", fmt.Sprintf("%d", expiredAt)).Result()
	if err != nil {
		return fmt.Errorf("failed to cleanup expired services: %w", err)
	}

	if removed > 0 {
		fmt.Printf("Cleaned up %d expired services for %s\n", removed, serviceName)
	}

	return nil
}

// Discover 发现服务（返回所有健康的服务实例）
func (d *RedisServiceDiscovery) Discover(ctx context.Context, serviceName string) ([]*ServiceInfo, error) {
	zsetKey := fmt.Sprintf("%s:%s:zset", d.prefix, serviceName)

	// 先清理过期服务（TTL设为30秒，给一点缓冲）
	if err := d.cleanupExpiredServices(ctx, serviceName, 30*time.Second); err != nil {
		fmt.Printf("Warning: cleanup failed: %v\n", err)
	}

	// 获取所有成员
	members, err := d.client.ZRange(ctx, zsetKey, 0, -1).Result()
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
