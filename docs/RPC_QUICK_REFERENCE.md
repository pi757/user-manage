# RPC编解码器快速参考

## 一句话总结

**默认使用MessagePack（更快更小），可选JSON（可读性好）**

## 快速开始

### 默认配置（推荐）✅

```go
// TCP Server - 自动使用MessagePack
server := rpc.NewServer()

// HTTP Server - 自动使用MessagePack  
pool := rpc.NewClientPool("localhost:9090", 100)
```

### 使用JSON

```go
// TCP Server - 显式指定JSON
server := rpc.NewServer(rpc.WithCodecType(rpc.JSONCodec))

// HTTP Server - 显式指定JSON
pool := rpc.NewClientPool("localhost:9090", 100, 
    rpc.WithPoolCodecType(rpc.JSONCodec))
```

## 选择指南

| 场景 | 推荐 | 原因 |
|------|------|------|
| 生产环境 | MessagePack | 性能最优 |
| 高并发 | MessagePack | 降低延迟 |
| 开发调试 | JSON | 便于查看 |
| 跨语言对接 | 根据对方选择 | 兼容性 |
| 日志记录 | JSON | 可读性好 |

## 关键规则

### ✅ 正确做法

```go
// 服务器和客户端使用相同编解码器
server := rpc.NewServer()  // MsgPack
pool := rpc.NewClientPool(addr, 100)  // MsgPack ✓
```

### ❌ 错误做法

```go
// 编解码器不匹配 - 会导致通信失败！
server := rpc.NewServer()  // MsgPack
pool := rpc.NewClientPool(addr, 100, 
    rpc.WithPoolCodecType(rpc.JSONCodec))  // JSON ✗
```

## 性能数据

```
MessagePack vs JSON:
- 序列化速度: 快 30-40%
- 反序列化速度: 快 40-50%  
- 数据大小: 小 30-50%
- 网络带宽: 节省 30-40%
```

## 常用API

```go
// 编解码器类型
rpc.MsgPackCodec  // MessagePack (默认)
rpc.JSONCodec     // JSON

// 配置选项
rpc.WithCodecType(codecType)           // Server
rpc.WithClientCodecType(codecType)     // Client
rpc.WithPoolCodecType(codecType)       // Pool
```

## 测试命令

```bash
# 运行单元测试
go test -v ./rpc

# 运行性能对比
go test -v ./rpc -run TestPerformanceComparison

# 运行基准测试
go test -bench=. ./rpc -benchmem
```

## 故障排查

### 问题: 连接成功但调用失败

**可能原因**: 编解码器不匹配

**解决方案**:
```bash
# 检查Server配置
grep -n "NewServer" cmd/tcp-server/main.go

# 检查Client配置  
grep -n "NewClientPool" cmd/http-server/main.go

# 确保两者一致
```

### 问题: 想查看RPC原始数据

**临时方案**: 切换到JSON
```go
server := rpc.NewServer(rpc.WithCodecType(rpc.JSONCodec))
```

**永久方案**: 添加日志中间件（待实现）

## 更多信息

- 📖 详细文档: `docs/RPC_CODEC_USAGE.md`
- 📊 优化总结: `docs/RPC_CODEC_OPTIMIZATION_SUMMARY.md`
- 🧪 测试代码: `rpc/codec_benchmark_test.go`

---

**提示**: 99%的场景使用默认配置即可，无需手动指定！
