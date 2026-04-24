package rpc

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

// CodecType 编解码器类型
type CodecType int

const (
	// MsgPackCodec MessagePack编解码器(默认,高效)
	MsgPackCodec CodecType = iota
	// JSONCodec JSON编解码器(兼容性好)
	JSONCodec
)

// Request RPC请求
type Request struct {
	ID     uint64                 `msgpack:"id" json:"id"`
	Method string                 `msgpack:"method" json:"method"`
	Params map[string]interface{} `msgpack:"params" json:"params"`
}

// Response RPC响应
type Response struct {
	ID     uint64      `msgpack:"id" json:"id"`
	Result interface{} `msgpack:"result,omitempty" json:"result,omitempty"`
	Error  string      `msgpack:"error,omitempty" json:"error,omitempty"`
}

// Handler RPC方法处理器
type Handler func(params map[string]interface{}) (interface{}, error)

// Server RPC服务器
type Server struct {
	listener  net.Listener
	handlers  map[string]Handler
	codecType CodecType
	mu        sync.RWMutex
}

// ServerOption 服务器配置选项
type ServerOption func(*Server)

// NewServer 创建RPC服务器
func NewServer(opts ...ServerOption) *Server {
	s := &Server{
		handlers:  make(map[string]Handler),
		codecType: MsgPackCodec, // 默认使用MessagePack
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
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
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	for {
		var req Request
		if err := s.decodeRequest(conn, &req); err != nil {
			if err == io.EOF {
				return
			}
			fmt.Printf("failed to decode request: %v\n", err)
			return
		}

		resp := s.handleRequest(&req)
		if err := s.encodeResponse(conn, resp); err != nil {
			fmt.Printf("failed to encode response: %v\n", err)
			return
		}
	}
}

// decodeRequest 解码请求(根据配置的编解码器类型)
func (s *Server) decodeRequest(conn net.Conn, req *Request) error {
	switch s.codecType {
	case MsgPackCodec:
		decoder := msgpack.NewDecoder(conn)
		return decoder.Decode(req)
	case JSONCodec:
		decoder := json.NewDecoder(conn)
		return decoder.Decode(req)
	default:
		decoder := msgpack.NewDecoder(conn)
		return decoder.Decode(req)
	}
}

// encodeResponse 编码响应(根据配置的编解码器类型)
func (s *Server) encodeResponse(conn net.Conn, resp *Response) error {
	switch s.codecType {
	case MsgPackCodec:
		encoder := msgpack.NewEncoder(conn)
		return encoder.Encode(resp)
	case JSONCodec:
		encoder := json.NewEncoder(conn)
		return encoder.Encode(resp)
	default:
		encoder := msgpack.NewEncoder(conn)
		return encoder.Encode(resp)
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
	conn      net.Conn
	codecType CodecType
	mu        sync.Mutex
}

// ClientOption 客户端配置选项
type ClientOption func(*Client)

// WithClientCodecType 设置客户端编解码器类型
func WithClientCodecType(codecType CodecType) ClientOption {
	return func(c *Client) {
		c.codecType = codecType
	}
}

// NewClient 创建RPC客户端
func NewClient(addr string, opts ...ClientOption) (*Client, error) {
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	client := &Client{
		conn:      conn,
		codecType: MsgPackCodec, // 默认使用MessagePack
	}

	for _, opt := range opts {
		opt(client)
	}

	return client, nil
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

	if err := c.encodeRequest(req); err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	var resp Response
	if err := c.decodeResponse(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("rpc error: %s", resp.Error)
	}

	return &resp, nil
}

// encodeRequest 编码请求(根据配置的编解码器类型)
func (c *Client) encodeRequest(req Request) error {
	switch c.codecType {
	case MsgPackCodec:
		encoder := msgpack.NewEncoder(c.conn)
		return encoder.Encode(req)
	case JSONCodec:
		encoder := json.NewEncoder(c.conn)
		return encoder.Encode(req)
	default:
		encoder := msgpack.NewEncoder(c.conn)
		return encoder.Encode(req)
	}
}

// decodeResponse 解码响应(根据配置的编解码器类型)
func (c *Client) decodeResponse(resp *Response) error {
	switch c.codecType {
	case MsgPackCodec:
		decoder := msgpack.NewDecoder(c.conn)
		return decoder.Decode(resp)
	case JSONCodec:
		decoder := json.NewDecoder(c.conn)
		return decoder.Decode(resp)
	default:
		decoder := msgpack.NewDecoder(c.conn)
		return decoder.Decode(resp)
	}
}

// Close 关闭客户端连接
func (c *Client) Close() error {
	return c.conn.Close()
}
