package rpc

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ClientPool RPC客户端连接池
type ClientPool struct {
	addr    string
	pool    chan *Client
	maxSize int
	mu      sync.Mutex
}

// NewClientPool 创建客户端连接池
func NewClientPool(addr string, maxSize int) *ClientPool {
	return &ClientPool{
		addr:    addr,
		pool:    make(chan *Client, maxSize),
		maxSize: maxSize,
	}
}

// Get 从连接池获取客户端
func (p *ClientPool) Get() (*Client, error) {
	select {
	case client := <-p.pool:
		// 检查连接是否仍然有效
		if client == nil || client.conn == nil {
			return NewClient(p.addr)
		}
		// 尝试设置读超时来检测连接状态
		if err := client.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
			client.Close()
			return NewClient(p.addr)
		}
		// 重置超时
		client.conn.SetReadDeadline(time.Time{})
		return client, nil
	default:
		return NewClient(p.addr)
	}
}

// Put 将客户端放回连接池
func (p *ClientPool) Put(client *Client) {
	select {
	case p.pool <- client:
	default:
		client.Close()
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
		client.Close()
	}
}

// InvokeHelper RPC调用辅助函数
func InvokeHelper(pool *ClientPool, method string, params map[string]interface{}, result interface{}) error {
	resp, err := pool.CallWithPool(method, params)
	if err != nil {
		return err
	}

	if resp.Result == nil {
		return nil
	}

	// 将结果转换为map,再序列化/反序列化为目标类型
	resultBytes, err := json.Marshal(resp.Result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := json.Unmarshal(resultBytes, result); err != nil {
		return fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return nil
}
