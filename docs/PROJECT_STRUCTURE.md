# 项目文件结构说明

## 目录结构

```
user-management-system/
│
├── cmd/                          # 应用入口目录
│   ├── tcp-server/              # TCP服务器主程序
│   │   └── main.go              # RPC服务器启动入口
│   ├── http-server/             # HTTP服务器主程序
│   │   └── main.go              # Gin Web服务器启动入口
│   ├── seed/                    # 数据种子程序
│   │   └── main.go              # 生成1000万测试数据
│   └── benchmark/               # 性能测试程序
│       └── main.go              # 压力测试工具
│
├── config/                      # 配置管理
│   └── config.go                # 配置加载和环境变量处理
│
├── models/                      # 数据模型
│   └── user.go                  # User和Session模型定义
│
├── database/                    # 数据库连接层
│   ├── mysql.go                 # MySQL初始化和连接池
│   └── redis.go                 # Redis初始化和连接池
│
├── rpc/                         # 自研RPC框架
│   ├── rpc.go                   # RPC核心实现(Server + Client)
│   ├── pool.go                  # RPC客户端连接池
│   └── rpc_test.go              # RPC单元测试
│
├── auth/                        # 认证授权模块
│   ├── auth.go                  # Session管理和密码加密
│   └── auth_test.go             # 认证模块测试
│
├── service/                     # 业务逻辑层
│   └── user_service.go          # 用户相关业务逻辑
│
├── handlers/                    # HTTP处理器层
│   ├── http_handler.go          # HTTP请求处理(登录、资料等)
│   ├── upload_handler.go        # 文件上传处理
│   └── http_handler_test.go     # HTTP处理器测试
│
├── web/                         # 前端页面
│   ├── login.html               # 登录页面
│   └── profile.html             # 个人资料页面
│
├── uploads/                     # 文件上传存储目录
│   └── .gitkeep                 # Git占位文件
│
├── docs/                        # 文档目录
│   ├── TECHNICAL_DESIGN.md      # 技术设计方案
│   ├── DEPLOYMENT.md            # 部署运维文档
│   ├── PERFORMANCE_TEST_REPORT.md  # 性能测试报告
│   └── SUMMARY.md               # 项目总结文档
│
├── docker-compose.yml           # Docker编排配置
├── Dockerfile.tcp               # TCP Server Docker镜像
├── Dockerfile.http              # HTTP Server Docker镜像
├── start.sh                     # 快速启动脚本
├── go.mod                       # Go模块依赖
├── go.sum                       # 依赖校验文件
├── .gitignore                   # Git忽略配置
├── README.md                    # 项目说明文档
└── main.go                      # 项目入口提示
```

## 核心文件说明

### 1. 应用入口 (cmd/)

#### cmd/tcp-server/main.go
- **功能**: 启动TCP RPC服务器
- **职责**: 
  - 初始化数据库连接
  - 注册RPC方法
  - 监听TCP端口
  - 处理RPC请求

#### cmd/http-server/main.go
- **功能**: 启动HTTP Web服务器
- **职责**:
  - 创建Gin路由
  - 注册HTTP端点
  - 通过RPC调用TCP Server
  - 提供静态文件服务

#### cmd/seed/main.go
- **功能**: 生成测试数据
- **特点**:
  - 并发插入(10个goroutine)
  - 批量操作(每批1000条)
  - 支持1000万用户数据

#### cmd/benchmark/main.go
- **功能**: 性能压力测试
- **测试场景**:
  - 200/2000并发
  - 固定/随机用户
  - QPS统计

### 2. 基础设施层 (database/, config/)

#### database/mysql.go
- GORM初始化
- 连接池配置
- 自动迁移

#### database/redis.go
- Redis客户端初始化
- 连接池配置
- 健康检查

#### config/config.go
- 配置结构体定义
- 环境变量加载
- 默认值设置

### 3. 通信层 (rpc/)

#### rpc/rpc.go
**Server端:**
- TCP监听
- JSON序列化/反序列化
- 方法注册和路由
- 并发处理

**Client端:**
- TCP连接
- 请求发送
- 响应接收
- 错误处理

#### rpc/pool.go
- 连接池管理
- 连接复用
- 自动扩缩容

### 4. 业务层 (auth/, service/)

#### auth/auth.go
**核心功能:**
- Token生成(crypto/rand)
- 密码哈希(bcrypt)
- Session创建/验证/删除
- Redis + MySQL双写

#### service/user_service.go
**业务方法:**
- Login: 用户登录
- GetProfile: 获取资料
- UpdateProfile: 更新资料
- Logout: 用户登出
- ValidateToken: Token验证

### 5. 表现层 (handlers/, web/)

#### handlers/http_handler.go
**HTTP端点:**
- POST /api/login
- POST /api/logout
- GET /api/profile
- PUT /api/profile

#### handlers/upload_handler.go
**文件上传:**
- POST /api/upload/avatar
- 文件类型验证
- 文件大小限制
- 头像URL更新

#### web/*.html
- 现代化UI设计
- 响应式布局
- AJAX异步请求
- Token本地存储

## 数据流向

### 登录流程
```
Browser → HTTP Server → RPC → TCP Server → MySQL(查询用户)
                                             ↓
                                         bcrypt验证
                                             ↓
                                         创建Session
                                             ↓
                                    Redis + MySQL存储
                                             ↓
Browser ← HTTP Server ← RPC ← 返回Token
```

### 获取资料流程
```
Browser → HTTP Server (携带Token)
             ↓
         RPC → TCP Server
                  ↓
              Redis验证Token
                  ↓
              MySQL查询用户
                  ↓
Browser ← 返回用户信息
```

### 更新资料流程
```
Browser → HTTP Server (Token + 新数据)
             ↓
         RPC → TCP Server
                  ↓
              Redis验证Token
                  ↓
              MySQL更新记录
                  ↓
Browser ← 返回更新结果
```

## 关键技术点

### 1. RPC协议设计
```json
Request:  {"id": 123, "method": "user.login", "params": {...}}
Response: {"id": 123, "result": {...}, "error": null}
```

### 2. Session管理
- Token: 64字符随机字符串
- 存储: Redis(主) + MySQL(备)
- TTL: 3600秒
- 刷新: 每次访问自动续期

### 3. 密码安全
- 算法: bcrypt
- Cost: 10
- 盐值: 自动生成

### 4. 性能优化
- 连接池: RPC(100-300), MySQL(100-200)
- 缓存: Redis Session
- 索引: username, token, deleted_at
- 并发: Goroutine池

## 扩展指南

### 添加新的RPC方法

1. 在`service/user_service.go`中实现业务逻辑
```go
func (s *UserService) NewMethod(params map[string]interface{}) (interface{}, error) {
    // 业务逻辑
}
```

2. 在`cmd/tcp-server/main.go`中注册
```go
rpcServer.Register("user.newMethod", func(params map[string]interface{}) (interface{}, error) {
    return userService.NewMethod(params)
})
```

3. 在HTTP Handler中调用
```go
resp, err := h.rpcPool.CallWithPool("user.newMethod", params)
```

### 添加新的HTTP端点

1. 在`handlers/http_handler.go`中添加处理方法
```go
func (h *Handler) NewEndpoint(c *gin.Context) {
    // 处理逻辑
}
```

2. 在`cmd/http-server/main.go`中注册路由
```go
api.POST("/new-endpoint", handler.NewEndpoint)
```

### 添加新的数据模型

1. 在`models/`中定义结构体
```go
type NewModel struct {
    ID   uint   `gorm:"primarykey"`
    Name string `gorm:"size:128"`
}
```

2. 在`database/mysql.go`中注册迁移
```go
DB.AutoMigrate(&models.NewModel{})
```

## 开发规范

### 代码风格
- 遵循Effective Go
- 使用golint检查
- 使用go vet检查
- 函数注释完整

### 提交规范
- feat: 新功能
- fix: 修复bug
- docs: 文档更新
- perf: 性能优化
- test: 测试相关

### 测试要求
- 核心逻辑必须有单元测试
- 性能指标必须达标
- 通过golint和go vet

---

**最后更新**: 2024-01-01
