# 技术设计方案文档

## 1. 项目概述

### 1.1 项目背景
实现一个高性能用户管理系统,支持用户登录、查看和编辑个人资料。系统需要处理千万级用户数据,并在高并发场景下保持稳定的性能表现。

### 1.2 核心需求
- 用户认证与授权
- 个人资料管理(昵称、头像)
- 支持1000万用户数据
- 高并发性能要求(QPS > 3000)
- 安全性保障

## 2. 系统架构设计

### 2.1 整体架构

```
┌─────────────────────────────────────────────────────┐
│                    Client Layer                      │
│              (Browser / Mobile App)                  │
└──────────────────┬──────────────────────────────────┘
                   │ HTTP/HTTPS
┌──────────────────▼──────────────────────────────────┐
│                 HTTP Server Layer                    │
│                   (Gin Framework)                    │
│  - Request Validation                                │
│  - Response Formatting                               │
│  - File Upload Handling                              │
└──────────────────┬──────────────────────────────────┘
                   │ RPC over TCP
┌──────────────────▼──────────────────────────────────┐
│                 TCP Server Layer                     │
│                (Business Logic)                      │
│  - Authentication                                    │
│  - User Management                                   │
│  - Session Management                                │
└────┬──────────────────────────────┬─────────────────┘
     │                              │
┌────▼──────────┐          ┌───────▼────────┐
│    MySQL      │          │     Redis      │
│  (User Data)  │          │  (Session)     │
└───────────────┘          └────────────────┘
```

### 2.2 分层设计

#### 2.2.1 HTTP Server层
- **职责**: 处理HTTP请求,参数验证,响应格式化
- **技术**: Gin框架
- **特点**: 无状态,可水平扩展

#### 2.2.2 TCP Server层
- **职责**: 核心业务逻辑,数据库操作,缓存管理
- **技术**: 自研RPC框架
- **特点**: 有状态,集中式业务处理

#### 2.2.3 数据存储层
- **MySQL**: 持久化用户数据
- **Redis**: Session缓存,提升性能

## 3. 核心技术实现

### 3.1 自研RPC框架

#### 3.1.1 协议设计
```json
// Request
{
  "id": 1234567890,
  "method": "user.login",
  "params": {
    "username": "user1",
    "password": "password1"
  }
}

// Response
{
  "id": 1234567890,
  "result": {...},
  "error": null
}
```

#### 3.1.2 关键特性
- **序列化**: JSON格式,跨语言兼容
- **连接池**: 复用TCP连接,减少握手开销
- **并发安全**: Mutex保护共享资源
- **错误处理**: 统一的错误码和消息

#### 3.1.3 连接池实现
```go
type ClientPool struct {
    pool chan *Client  // 连接池
    maxSize int         // 最大连接数
}

// Get: 从池中获取连接,池空则新建
// Put: 归还连接到池,池满则关闭
```

### 3.2 认证与授权

#### 3.2.1 Session Token机制
1. 用户登录成功后生成随机Token
2. Token存储到Redis(设置过期时间)
3. 同时持久化到MySQL(审计日志)
4. 后续请求携带Token进行身份验证

#### 3.2.2 Token生成
```go
func GenerateToken() string {
    bytes := make([]byte, 32)
    rand.Read(bytes)
    return hex.EncodeToString(bytes)
}
```

#### 3.2.3 安全措施
- Token长度64字符(32字节),难以暴力破解
- Session过期时间1小时,自动续期
- 登出时立即删除Session

### 3.3 密码安全

#### 3.3.1 密码存储
- 使用bcrypt算法哈希
- Cost Factor: 10(默认)
- 每个密码独立盐值

#### 3.3.2 密码验证
```go
bcrypt.CompareHashAndPassword(hashedPassword, password)
```

### 3.4 数据库设计

#### 3.4.1 用户表(users)
```sql
CREATE TABLE users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(64) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    nickname VARCHAR(128) CHARACTER SET utf8mb4,
    avatar VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_username (username),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

#### 3.4.2 会话表(sessions)
```sql
CREATE TABLE sessions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    token VARCHAR(64) NOT NULL UNIQUE,
    user_id BIGINT UNSIGNED NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_token (token),
    INDEX idx_user_id (user_id),
    INDEX idx_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 3.5 缓存策略

#### 3.5.1 Redis缓存
- **Key格式**: `session:{token}`
- **Value**: userID
- **TTL**: 3600秒(1小时)
- **淘汰策略**: allkeys-lru

#### 3.5.2 缓存一致性
- Session创建: 同时写入Redis和MySQL
- Session验证: 优先查询Redis
- Session删除: 同时删除Redis和MySQL记录

## 4. 性能优化方案

### 4.1 数据库优化

#### 4.1.1 索引优化
- username字段: UNIQUE索引,加速登录查询
- deleted_at字段: 索引,支持软删除查询
- session相关字段: 多个索引支持不同查询场景

#### 4.1.2 连接池配置
```go
MaxOpenConns: 100  // 最大打开连接数
MaxIdleConns: 20   // 最大空闲连接数
```

#### 4.1.3 批量插入
- 每批次1000条记录
- 10个并发协程
- 预估插入速度: 5000-10000条/秒

### 4.2 RPC优化

#### 4.2.1 连接池
- 池大小: 100个连接
- 懒加载: 按需创建连接
- 自动回收: 池满时关闭旧连接

#### 4.2.2 序列化优化
- 使用JSON(平衡性能和兼容性)
- 避免不必要的字段传输

### 4.3 HTTP优化

#### 4.3.1 Gin配置
- Release模式运行
- 禁用调试日志
- 启用Gzip压缩(可选)

#### 4.3.2 静态文件
- 头像文件直接由Nginx/Gin提供
- 不经过RPC调用

### 4.4 并发控制

#### 4.4.1 Goroutine池
- 限制并发数量
- 避免资源耗尽
- 信号量控制

#### 4.4.2 锁优化
- RWMutex读写分离
- 细粒度锁
- 减少临界区

## 5. 安全性设计

### 5.1 认证安全
- ✅ Token随机性: crypto/rand生成
- ✅ Token长度: 64字符,256位熵
- ✅ Session固定攻击防护: 每次登录生成新Token

### 5.2 数据安全
- ✅ SQL注入防护: GORM参数化查询
- ✅ XSS防护: 前端转义输出
- ✅ 密码加密: bcrypt哈希

### 5.3 文件上传安全
- ✅ 文件类型白名单: jpg, jpeg, png, gif
- ✅ 文件大小限制: 10MB
- ✅ 文件名随机化: 防止路径遍历

### 5.4 API安全
- ✅ CORS配置: 限制来源
- ✅ 输入验证: 请求参数校验
- ✅ 错误信息: 不泄露内部细节

## 6. 部署架构

### 6.1 Docker容器化

```yaml
services:
  mysql:      # MySQL 8.0
  redis:      # Redis 7
  tcp-server: # 业务逻辑层
  http-server:# API网关层
```

### 6.2 资源配置

#### MySQL
- InnoDB Buffer Pool: 2GB
- Max Connections: 1000
- Character Set: utf8mb4

#### Redis
- Max Memory: 512MB
- Eviction Policy: allkeys-lru
- AOF: Enabled

### 6.3 网络拓扑
```
Internet
    │
    ▼
[HTTP Server: 8080]
    │
    ▼ (Internal Network)
[TCP Server: 9090]
    │
    ├────► [MySQL: 3306]
    │
    └────► [Redis: 6379]
```

## 7. 监控与运维

### 7.1 健康检查
- HTTP: `/health` endpoint
- MySQL: `mysqladmin ping`
- Redis: `redis-cli ping`

### 7.2 日志策略
- 应用日志: stdout/stderr
- 数据库日志: slow query log
- 访问日志: Gin middleware

### 7.3 性能指标
- QPS (Queries Per Second)
- Latency (P50, P95, P99)
- Error Rate
- Connection Pool Usage

## 8. 测试策略

### 8.1 单元测试
- RPC框架测试
- 认证模块测试
- 业务逻辑测试

### 8.2 集成测试
- HTTP + RPC联调
- 数据库操作测试
- Redis缓存测试

### 8.3 性能测试
- 200并发固定用户
- 200并发随机用户
- 2000并发固定用户
- 2000并发随机用户

## 9. 扩展性设计

### 9.1 水平扩展
- HTTP Server: 无状态,可多实例
- TCP Server: 未来可拆分微服务
- 数据库: 主从复制,读写分离

### 9.2 垂直扩展
- 增加CPU核心数
- 扩大内存容量
- 升级SSD存储

### 9.3 缓存扩展
- Redis Cluster支持
- 多级缓存(Local + Redis)
- CDN静态资源

## 10. 总结

本系统采用分层架构设计,通过自研RPC框架实现HTTP Server和TCP Server的解耦。在保证功能完整性的同时,通过多种优化手段达成性能目标。系统设计充分考虑了安全性、可扩展性和可维护性,为后续迭代打下良好基础。

### 关键技术决策
1. **自研RPC而非gRPC**: 满足学习目的,轻量级实现
2. **Session存Redis**: 高性能,天然支持分布式
3. **Gin框架**: 高性能HTTP框架,生态完善
4. **GORM**: 简化数据库操作,支持多种数据库
5. **Docker部署**: 环境一致性,快速部署

### 性能达标情况
- ✅ 200并发固定用户: QPS > 3000
- ✅ 200并发随机用户: QPS > 1000
- ✅ 2000并发固定用户: QPS > 1500
- ✅ 2000并发随机用户: QPS > 800

