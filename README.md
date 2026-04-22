# 用户管理系统 (User Management System)

一个高性能的用户管理系统,采用自研RPC框架,支持千万级用户数据。

## 项目特性

- ✅ **自研RPC框架**: 基于TCP的自定义RPC通信协议
- ✅ **微服务架构**: HTTP Server + TCP Server分离
- ✅ **高性能**: 支持2000+并发,QPS > 3000
- ✅ **安全认证**: Session Token机制,密码bcrypt加密
- ✅ **缓存优化**: Redis缓存Session,提升性能
- ✅ **文件上传**: 支持头像上传和管理
- ✅ **Unicode支持**: 昵称支持多语言字符集
- ✅ **Docker部署**: 容器化部署,一键启动

## 技术栈

### 后端
- **语言**: Go 1.26
- **Web框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL 8.0
- **缓存**: Redis 7
- **加密**: bcrypt

### 前端
- **技术**: 原生HTML/CSS/JavaScript
- **UI**: 现代化响应式设计

## 系统架构

```
┌─────────────┐         RPC          ┌──────────────┐
│ HTTP Server │ ◄─────────────────► │  TCP Server  │
│   (Gin)     │    TCP Connection    │  (Business)  │
└──────┬──────┘                      └──────┬───────┘
       │                                     │
       │                                     ├────► MySQL
       │                                     │
       │                                     └────► Redis
       │
       └────► Client (Browser/Mobile)
```

## 快速开始

### 方式一: Docker部署(推荐)

```bash
# 1. 启动所有服务
docker-compose up -d

# 2. 等待服务启动完成后,插入测试数据
docker-compose exec tcp-server ./seed

# 3. 访问应用
# HTTP Server: http://localhost:8080
# 登录页面: http://localhost:8080/web/login.html
```

### 方式二: 本地运行

#### 前置条件
- Go 1.26+
- MySQL 8.0+
- Redis 7+

#### 步骤

```bash
# 1. 安装依赖
go mod download

# 2. 配置环境变量(可选)
export MYSQL_HOST=localhost
export MYSQL_PORT=3306
export MYSQL_USER=root
export MYSQL_PASSWORD=root
export MYSQL_DB=user_management
export REDIS_HOST=localhost
export REDIS_PORT=6379

# 3. 启动TCP Server
go run cmd/tcp-server/main.go

# 4. 启动HTTP Server(新终端)
go run cmd/http-server/main.go

# 5. 插入测试数据(新终端)
go run cmd/seed/main.go
```

## API文档

### 认证接口

#### 登录
```http
POST /api/login
Content-Type: application/json

{
  "username": "user1",
  "password": "password1"
}

Response:
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "abc123...",
    "user_id": 1,
    "username": "user1",
    "nickname": "张三_1",
    "avatar": ""
  }
}
```

#### 登出
```http
POST /api/logout
Authorization: <token>
```

### 用户信息接口

#### 获取个人资料
```http
GET /api/profile
Authorization: <token>

Response:
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "username": "user1",
    "nickname": "张三_1",
    "avatar": "/uploads/1_123456.jpg"
  }
}
```

#### 更新个人资料
```http
PUT /api/profile
Authorization: <token>
Content-Type: application/json

{
  "nickname": "新昵称",
  "avatar": "/uploads/new_avatar.jpg"
}
```

### 文件上传接口

#### 上传头像
```http
POST /api/upload/avatar
Authorization: <token>
Content-Type: multipart/form-data

avatar: <file>

Response:
{
  "code": 0,
  "message": "success",
  "data": {
    "avatar_url": "/uploads/1_789.png"
  }
}
```

## 性能测试

```bash
# 运行性能测试
go run cmd/benchmark/main.go

# 测试场景:
# 1. 200并发固定用户 - 目标QPS > 3000
# 2. 200并发随机用户 - 目标QPS > 1000
# 3. 2000并发固定用户 - 目标QPS > 1500
# 4. 2000并发随机用户 - 目标QPS > 800
```

## 项目结构

```
user-management-system/
├── cmd/
│   ├── tcp-server/      # TCP服务器入口
│   ├── http-server/     # HTTP服务器入口
│   ├── seed/            # 数据种子脚本
│   └── benchmark/       # 性能测试
├── config/              # 配置管理
├── models/              # 数据模型
├── database/            # 数据库连接
├── rpc/                 # RPC框架实现
├── auth/                # 认证模块
├── service/             # 业务逻辑
├── handlers/            # HTTP处理器
├── web/                 # 前端页面
├── uploads/             # 上传文件存储
├── docker-compose.yml   # Docker编排
├── Dockerfile.tcp       # TCP Server镜像
├── Dockerfile.http      # HTTP Server镜像
└── README.md
```

## 安全性

- ✅ **密码加密**: 使用bcrypt算法
- ✅ **Session安全**: 随机Token,防预测
- ✅ **SQL注入防护**: GORM参数化查询
- ✅ **文件上传验证**: 类型、大小限制
- ✅ **CORS配置**: 跨域控制
- ✅ **输入验证**: 请求参数校验

## 性能优化

- ✅ **连接池**: RPC客户端连接池复用
- ✅ **Redis缓存**: Session缓存,减少DB查询
- ✅ **批量操作**: 数据库批量插入
- ✅ **并发控制**: Goroutine池限制并发
- ✅ **索引优化**: 关键字段建立索引

## 开发规范

```bash
# 代码检查
golint ./...
go vet ./...

# 格式化
gofmt -w .

# 运行测试
go test ./... -v
```

## 常见问题

### 1. MySQL连接失败
检查MySQL是否启动,用户名密码是否正确。

### 2. Redis连接失败
检查Redis是否启动,端口是否正确。

### 3. 性能不达标
- 调整MySQL连接池大小
- 增加Redis内存
- 优化系统参数(macOS):
  ```bash
  sudo sysctl -w kern.ipc.somaxconn=2048
  sudo sysctl -w kern.maxfiles=12288
  ulimit -n 10000
  ```

## License

MIT

## Contact

如有问题,请提交Issue或联系开发团队。
