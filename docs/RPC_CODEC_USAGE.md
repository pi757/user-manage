# RPC编解码器使用指南

## 概述

RPC框架支持两种编解码器：
- **MessagePack**（默认）：二进制格式，更高效、更紧凑
- **JSON**：文本格式，更好的可读性和兼容性

## 性能对比

| 指标 | MessagePack | JSON |
|------|-------------|------|
| 序列化速度 | ⚡ 快 | 🐢 较慢 |
| 数据大小 | 📦 小30-50% | 📦 较大 |
| 可读性 | ❌ 二进制 | ✅ 文本 |
| 调试难度 | 🔧 较难 | 🔧 简单 |
| 跨语言支持 | ✅ 广泛 | ✅ 广泛 |

## 使用示例

### 1. TCP服务器配置

#### 使用MessagePack（推荐，默认）
```go
// cmd/tcp-server/main.go
rpcServer := rpc.NewServer() // 默认使用MessagePack
```

#### 使用JSON
```go
// cmd/tcp-server/main.go
rpcServer := rpc.NewServer(
    rpc.WithCodecType(rpc.JSONCodec),
)
```

### 2. HTTP服务器配置

#### 使用MessagePack（推荐，默认）
```go
// cmd/http-server/main.go
rpcPool := rpc.NewClientPool(100) // RPC框架内部管理服务器地址，默认使用MessagePack
```

#### 使用JSON
```go
// cmd/http-server/main.go
rpcPool := rpc.NewClientPool(
    100,
    rpc.WithPoolCodecType(rpc.JSONCodec),
)
```

### 3. 直接创建客户端

#### 使用MessagePack（推荐，默认）
```go
client, err := rpc.NewClient("localhost:9090") // 默认使用MessagePack
```

#### 使用JSON
```go
client, err := rpc.NewClient(
    "localhost:9090",
    rpc.WithClientCodecType(rpc.JSONCodec),
)
```

## 重要注意事项

### ⚠️ 编解码器必须匹配

**服务器和客户端必须使用相同的编解码器类型！**

```go
// ✅ 正确：都使用MessagePack
server := rpc.NewServer() // 默认MsgPackCodec
pool := rpc.NewClientPool(100) // 默认MsgPackCodec

// ✅ 正确：都使用JSON
server := rpc.NewServer(rpc.WithCodecType(rpc.JSONCodec))
pool := rpc.NewClientPool(100, rpc.WithPoolCodecType(rpc.JSONCodec))

// ❌ 错误：编解码器不匹配会导致通信失败
server := rpc.NewServer() // MsgPackCodec
pool := rpc.NewClientPool(100, rpc.WithPoolCodecType(rpc.JSONCodec))
```

### 迁移建议

如果现有系统已部署并使用JSON，可以逐步迁移：

1. **第一阶段**：保持JSON不变
2. **第二阶段**：在新服务中使用MessagePack测试
3. **第三阶段**：所有服务切换到MessagePack

## 技术细节

### Request/Response结构

```go
type Request struct {
    ID      uint64                 `msgpack:"id" json:"id"`
    Method  string                 `msgpack:"method" json:"method"`
    Params  map[string]interface{} `msgpack:"params" json:"params"`
}

type Response struct {
    ID     uint64      `msgpack:"id" json:"id"`
    Result interface{} `msgpack:"result,omitempty" json:"result,omitempty"`
    Error  string      `msgpack:"error,omitempty" json:"error,omitempty"`
}
```

结构体同时支持两种标签，确保兼容性。

### 底层实现

- **MessagePack**: 使用 `github.com/vmihailenco/msgpack/v5`
- **JSON**: 使用标准库 `encoding/json`

## 性能测试

运行以下命令查看实际的性能对比：

```bash
go test -v ./rpc -run TestCodecTypeComparison
```

典型结果：
```
JSON size: 87 bytes
MessagePack size: 58 bytes
Compression ratio: 66.67%
```

MessagePack通常能减少30-50%的数据传输量。

## 最佳实践

1. **生产环境**：使用MessagePack获得最佳性能
2. **开发调试**：可以使用JSON便于查看原始数据
3. **微服务间通信**：统一使用MessagePack
4. **对外API**：如果需要人类可读，使用JSON
5. **高并发场景**：务必使用MessagePack降低网络开销

## 常见问题

### Q: 如何判断当前使用的是哪种编解码器？

A: 检查创建Server或Client时的配置，未指定则默认为MessagePack。

### Q: 可以在运行时切换编解码器吗？

A: 不可以。编解码器在创建时确定，之后不能更改。需要重启服务。

### Q: MessagePack兼容所有数据类型吗？

A: 是的，MessagePack支持Go的所有基本类型和复杂类型（map、slice、struct等）。

### Q: 如果忘记设置编解码器会怎样？

A: 默认使用MessagePack，这是推荐的配置。
