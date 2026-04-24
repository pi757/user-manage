package rpc

import (
	"encoding/json"
	"testing"

	"github.com/vmihailenco/msgpack/v5"
)

func TestRequestResponse(t *testing.T) {
	// 测试JSON序列化
	req := Request{
		ID:     123,
		Method: "test.method",
		Params: map[string]interface{}{
			"key": "value",
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request with JSON: %v", err)
	}

	var decodedReq Request
	err = json.Unmarshal(data, &decodedReq)
	if err != nil {
		t.Fatalf("Failed to unmarshal request with JSON: %v", err)
	}

	if decodedReq.ID != req.ID {
		t.Errorf("Expected ID %d, got %d", req.ID, decodedReq.ID)
	}

	if decodedReq.Method != req.Method {
		t.Errorf("Expected Method %s, got %s", req.Method, decodedReq.Method)
	}
}

func TestRequestResponseMsgPack(t *testing.T) {
	// 测试MessagePack序列化
	req := Request{
		ID:     456,
		Method: "test.msgpack",
		Params: map[string]interface{}{
			"name": "test",
			"age":  25,
		},
	}

	data, err := msgpack.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request with MessagePack: %v", err)
	}

	var decodedReq Request
	err = msgpack.Unmarshal(data, &decodedReq)
	if err != nil {
		t.Fatalf("Failed to unmarshal request with MessagePack: %v", err)
	}

	if decodedReq.ID != req.ID {
		t.Errorf("Expected ID %d, got %d", req.ID, decodedReq.ID)
	}

	if decodedReq.Method != req.Method {
		t.Errorf("Expected Method %s, got %s", req.Method, decodedReq.Method)
	}
}

func TestCodecTypeComparison(t *testing.T) {
	// 比较JSON和MessagePack的性能和大小
	req := Request{
		ID:     789,
		Method: "user.login",
		Params: map[string]interface{}{
			"username": "testuser",
			"password": "testpass123",
		},
	}

	// JSON序列化
	jsonData, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal with JSON: %v", err)
	}

	// MessagePack序列化
	msgpackData, err := msgpack.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal with MessagePack: %v", err)
	}

	// 比较大小
	t.Logf("JSON size: %d bytes", len(jsonData))
	t.Logf("MessagePack size: %d bytes", len(msgpackData))
	t.Logf("Compression ratio: %.2f%%", float64(len(msgpackData))/float64(len(jsonData))*100)

	if len(msgpackData) >= len(jsonData) {
		t.Logf("Warning: MessagePack is not smaller than JSON in this case")
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
