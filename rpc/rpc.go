package rpc

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// Request RPC请求
type Request struct {
	ID      uint64                 `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
}

// Response RPC响应
type Response struct {
	ID     uint64      `json:"id"`
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// Handler RPC方法处理器
type Handler func(params map[string]interface{}) (interface{}, error)

// Server RPC服务器
type Server struct {
	listener net.Listener
	handlers map[string]Handler
	mu       sync.RWMutex
}

// NewServer 创建RPC服务器
func NewServer() *Server {
	return &Server{
		handlers: make(map[string]Handler),
	}
}

// Register 注册RPC方法
func (s *Server) Register(method string, handler Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[method] = handler
}

// Start 启动RPC服务器
func (s *Server) Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = listener

	fmt.Printf("RPC server started on %s\n", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("failed to accept connection: %v\n", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

// handleConnection 处理单个连接
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		var req Request
		if err := decoder.Decode(&req); err != nil {
			if err == io.EOF {
				return
			}
			fmt.Printf("failed to decode request: %v\n", err)
			return
		}

		resp := s.handleRequest(&req)
		if err := encoder.Encode(resp); err != nil {
			fmt.Printf("failed to encode response: %v\n", err)
			return
		}
	}
}

// handleRequest 处理单个请求
func (s *Server) handleRequest(req *Request) *Response {
	s.mu.RLock()
	handler, ok := s.handlers[req.Method]
	s.mu.RUnlock()

	if !ok {
		return &Response{
			ID:    req.ID,
			Error: fmt.Sprintf("method %s not found", req.Method),
		}
	}

	result, err := handler(req.Params)
	if err != nil {
		return &Response{
			ID:    req.ID,
			Error: err.Error(),
		}
	}

	return &Response{
		ID:     req.ID,
		Result: result,
	}
}

// Stop 停止RPC服务器
func (s *Server) Stop() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// Client RPC客户端
type Client struct {
	conn   net.Conn
	encoder *json.Encoder
	decoder *json.Decoder
	mu     sync.Mutex
}

// NewClient 创建RPC客户端
func NewClient(addr string) (*Client, error) {
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	return &Client{
		conn:    conn,
		encoder: json.NewEncoder(conn),
		decoder: json.NewDecoder(conn),
	}, nil
}

// Call 调用RPC方法
func (c *Client) Call(method string, params map[string]interface{}) (*Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	req := Request{
		ID:     uint64(time.Now().UnixNano()),
		Method: method,
		Params: params,
	}

	if err := c.encoder.Encode(req); err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	var resp Response
	if err := c.decoder.Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("rpc error: %s", resp.Error)
	}

	return &resp, nil
}

// Close 关闭客户端连接
func (c *Client) Close() error {
	return c.conn.Close()
}
