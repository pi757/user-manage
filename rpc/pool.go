package rpc

import (
	"fmt"
	"sync"
	"time"
)

// ClientPool RPC客户端连接池
type ClientPool struct {
	addr      string
	pool      chan *Client
	maxSize   int
	codecType CodecType
	mu        sync.Mutex
}

// PoolOption 连接池配置选项
type PoolOption func(*ClientPool)

// NewClientPool 创建客户端连接池
func NewClientPool(addr string, maxSize int, opts ...PoolOption) *ClientPool {
	p := &ClientPool{
		addr:      addr,
		pool:      make(chan *Client, maxSize),
		maxSize:   maxSize,
		codecType: MsgPackCodec, // 默认使用MessagePack
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Get 从连接池获取客户端
func (p *ClientPool) Get() (*Client, error) {
	select {
	case client := <-p.pool:
		// 检查连接是否仍然有效
		if client == nil || client.conn == nil {
			return NewClient(p.addr, WithClientCodecType(p.codecType))
		}
		// 尝试设置读超时来检测连接状态
		if err := client.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
			err := client.Close()
			if err != nil {
				return nil, err
			}
			return NewClient(p.addr, WithClientCodecType(p.codecType))
		}
		// 重置超时
		err := client.conn.SetReadDeadline(time.Time{})
		if err != nil {
			return nil, err
		}
		return client, nil
	default:
		return NewClient(p.addr, WithClientCodecType(p.codecType))
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
