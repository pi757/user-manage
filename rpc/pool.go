package rpc

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
	"user-management-system/discovery"
)

// DefaultServerAddr 默认RPC服务器地址（可通过环境变量RPC_SERVER_ADDR覆盖）
const DefaultServerAddr = ":9090"

// GetServerAddr 获取RPC服务器地址（优先使用环境变量）
func GetServerAddr() string {
	if addr := os.Getenv("RPC_SERVER_ADDR"); addr != "" {
		return addr
	}
	return DefaultServerAddr
}

// ClientPool RPC客户端连接池
type ClientPool struct {
	addr                string
	pool                chan *Client
	maxSize             int
	codecType           CodecType
	mu                  sync.Mutex
	serviceDiscovery    discovery.ServiceDiscovery
	serviceName         string
	useServiceDiscovery bool
}

// PoolOption 连接池配置选项
type PoolOption func(*ClientPool)

// WithPoolCodecType 设置连接池编解码器类型
func WithPoolCodecType(codecType CodecType) PoolOption {
	return func(p *ClientPool) {
		p.codecType = codecType
	}
}

// WithServiceDiscovery 使用服务发现（替代硬编码地址）
func WithServiceDiscovery(sd discovery.ServiceDiscovery, serviceName string) PoolOption {
	return func(p *ClientPool) {
		p.serviceDiscovery = sd
		p.serviceName = serviceName
		p.useServiceDiscovery = true
	}
}

// NewClientPool 创建客户端连接池（支持服务发现）
func NewClientPool(maxSize int, opts ...PoolOption) *ClientPool {
	p := &ClientPool{
		pool:      make(chan *Client, maxSize),
		maxSize:   maxSize,
		codecType: MsgPackCodec, // 默认使用MessagePack
	}

	for _, opt := range opts {
		opt(p)
	}

	// 如果未使用服务发现，则从环境变量获取地址
	if !p.useServiceDiscovery {
		p.addr = GetServerAddr()
	}

	return p
}

// resolveAddress 解析服务地址（优先使用服务发现）
func (p *ClientPool) resolveAddress() (string, error) {
	if p.useServiceDiscovery && p.serviceDiscovery != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		serviceInfo, err := p.serviceDiscovery.DiscoverOne(ctx, p.serviceName)
		if err != nil {
			return "", fmt.Errorf("failed to discover service %s: %w", p.serviceName, err)
		}

		return fmt.Sprintf("%s:%d", serviceInfo.Address, serviceInfo.Port), nil
	}

	// 使用配置的地址
	return p.addr, nil
}

// Get 从连接池获取客户端
func (p *ClientPool) Get() (*Client, error) {
	select {
	case client := <-p.pool:
		// 检查连接是否仍然有效
		if client == nil || client.conn == nil {
			addr, err := p.resolveAddress()
			if err != nil {
				return nil, err
			}
			return NewClient(addr, WithClientCodecType(p.codecType))
		}
		// 尝试设置读超时来检测连接状态
		if err := client.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
			err := client.Close()
			if err != nil {
				return nil, err
			}
			addr, err := p.resolveAddress()
			if err != nil {
				return nil, err
			}
			return NewClient(addr, WithClientCodecType(p.codecType))
		}
		// 重置超时
		err := client.conn.SetReadDeadline(time.Time{})
		if err != nil {
			return nil, err
		}
		return client, nil
	default:
		addr, err := p.resolveAddress()
		if err != nil {
			return nil, err
		}
		return NewClient(addr, WithClientCodecType(p.codecType))
	}
}

// Put 将客户端放回连接池
func (p *ClientPool) Put(client *Client) {
	select {
	case p.pool <- client:
	default:
		err := client.Close()
		if err != nil {
			return
		}
	}
}

// CallWithPool 使用连接池调用RPC
func (p *ClientPool) CallWithPool(method string, params map[string]interface{}) (*Response, error) {
	client, err := p.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get client from pool: %w", err)
	}

	resp, err := client.Call(method, params)
	if err != nil {
		// 连接失败，关闭旧连接并尝试重新获取
		client.Close()

		// 重试一次
		client, err = p.Get()
		if err != nil {
			return nil, fmt.Errorf("failed to get client from pool (retry): %w", err)
		}
		defer p.Put(client)

		resp, err = client.Call(method, params)
		if err != nil {
			client.Close()
			return nil, err
		}

		return resp, nil
	}

	// 成功则放回连接池
	p.Put(client)
	return resp, nil
}

// Close 关闭连接池
func (p *ClientPool) Close() {
	close(p.pool)
	for client := range p.pool {
		err := client.Close()
		if err != nil {
			return
		}
	}
}
