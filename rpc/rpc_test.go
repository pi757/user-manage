package rpc

import (
	"encoding/json"
	"testing"
)

func TestRequestResponse(t *testing.T) {
	// 测试序列化
	req := Request{
		ID:     123,
		Method: "test.method",
		Params: map[string]interface{}{
			"key": "value",
		},
	}
	
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}
	
	var decodedReq Request
	err = json.Unmarshal(data, &decodedReq)
	if err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}
	
	if decodedReq.ID != req.ID {
		t.Errorf("Expected ID %d, got %d", req.ID, decodedReq.ID)
	}
	
	if decodedReq.Method != req.Method {
		t.Errorf("Expected Method %s, got %s", req.Method, decodedReq.Method)
	}
}

func TestServerRegister(t *testing.T) {
	server := NewServer()
	
	// 注册一个测试方法
	server.Register("test.method", func(params map[string]interface{}) (interface{}, error) {
		return "result", nil
	})
	
	// 验证方法已注册
	server.mu.RLock()
	_, exists := server.handlers["test.method"]
	server.mu.RUnlock()
	
	if !exists {
		t.Error("Handler should be registered")
	}
}

func TestClientPool(t *testing.T) {
	// 注意: 这个测试需要真实的TCP Server
	// 在实际项目中应该使用mock
	t.Skip("Requires TCP server connection")
}
