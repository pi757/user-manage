package rpc

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

// BenchmarkJSONMarshal 测试JSON序列化性能
func BenchmarkJSONMarshal(b *testing.B) {
	req := Request{
		ID:     12345,
		Method: "user.login",
		Params: map[string]interface{}{
			"username": "testuser",
			"password": "testpass123",
			"email":    "test@example.com",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMsgPackMarshal 测试MessagePack序列化性能
func BenchmarkMsgPackMarshal(b *testing.B) {
	req := Request{
		ID:     12345,
		Method: "user.login",
		Params: map[string]interface{}{
			"username": "testuser",
			"password": "testpass123",
			"email":    "test@example.com",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := msgpack.Marshal(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkJSONUnmarshal 测试JSON反序列化性能
func BenchmarkJSONUnmarshal(b *testing.B) {
	req := Request{
		ID:     12345,
		Method: "user.login",
		Params: map[string]interface{}{
			"username": "testuser",
			"password": "testpass123",
			"email":    "test@example.com",
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decodedReq Request
		err := json.Unmarshal(data, &decodedReq)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMsgPackUnmarshal 测试MessagePack反序列化性能
func BenchmarkMsgPackUnmarshal(b *testing.B) {
	req := Request{
		ID:     12345,
		Method: "user.login",
		Params: map[string]interface{}{
			"username": "testuser",
			"password": "testpass123",
			"email":    "test@example.com",
		},
	}

	data, err := msgpack.Marshal(req)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decodedReq Request
		err := msgpack.Unmarshal(data, &decodedReq)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestPerformanceComparison 性能对比测试
func TestPerformanceComparison(t *testing.T) {
	req := Request{
		ID:     12345,
		Method: "user.login",
		Params: map[string]interface{}{
			"username": "testuser",
			"password": "testpass123",
			"email":    "test@example.com",
		},
	}

	// 测试JSON性能
	jsonStart := time.Now()
	jsonIterations := 100000
	for i := 0; i < jsonIterations; i++ {
		_, err := json.Marshal(req)
		if err != nil {
			return
		}
	}
	jsonDuration := time.Since(jsonStart)

	// 测试MessagePack性能
	msgpackStart := time.Now()
	msgpackIterations := 100000
	for i := 0; i < msgpackIterations; i++ {
		_, err := msgpack.Marshal(req)
		if err != nil {
			return
		}
	}
	msgpackDuration := time.Since(msgpackStart)

	// 测试数据大小
	jsonData, _ := json.Marshal(req)
	msgpackData, _ := msgpack.Marshal(req)

	// 输出结果
	fmt.Println("\n=== 编解码器性能对比 ===")
	fmt.Printf("迭代次数: %d\n", jsonIterations)
	fmt.Println()

	fmt.Println("序列化性能:")
	fmt.Printf("  JSON:        %v (%.2f ns/op)\n", jsonDuration, float64(jsonDuration)/float64(jsonIterations))
	fmt.Printf("  MessagePack: %v (%.2f ns/op)\n", msgpackDuration, float64(msgpackDuration)/float64(msgpackIterations))
	if jsonDuration > msgpackDuration {
		improvement := float64(jsonDuration-msgpackDuration) / float64(jsonDuration) * 100
		fmt.Printf("  提升:        %.2f%%\n", improvement)
	}
	fmt.Println()

	fmt.Println("数据大小:")
	fmt.Printf("  JSON:        %d bytes\n", len(jsonData))
	fmt.Printf("  MessagePack: %d bytes\n", len(msgpackData))
	reduction := float64(len(jsonData)-len(msgpackData)) / float64(len(jsonData)) * 100
	fmt.Printf("  减少:        %.2f%%\n", reduction)
	fmt.Println()

	fmt.Println("结论:")
	if msgpackDuration < jsonDuration && len(msgpackData) < len(jsonData) {
		fmt.Println("  ✅ MessagePack在性能和大小上都优于JSON")
	} else if msgpackDuration < jsonDuration {
		fmt.Println("  ⚡ MessagePack在性能上更优")
	} else if len(msgpackData) < len(jsonData) {
		fmt.Println("  📦 MessagePack在数据大小上更优")
	} else {
		fmt.Println("  ℹ️  两种编解码器性能相当")
	}
}
