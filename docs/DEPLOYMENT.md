# 部署与运维文档

## 1. 环境要求

### 1.1 硬件要求

#### 最低配置
- CPU: 4核心
- 内存: 8GB
- 磁盘: 50GB SSD
- 网络: 1Gbps

#### 推荐配置
- CPU: 8核心+
- 内存: 16GB+
- 磁盘: 100GB NVMe SSD
- 网络: 10Gbps

### 1.2 软件要求

#### Docker部署
- Docker: 20.10+
- Docker Compose: 2.0+

#### 本地部署
- Go: 1.26+
- MySQL: 8.0+
- Redis: 7.0+
- Linux/macOS/Windows

## 2. Docker部署(推荐)

### 2.1 快速启动

```bash
# 1. 克隆代码
git clone <repository-url>
cd user-management-system

# 2. 启动所有服务
docker-compose up -d

# 3. 查看服务状态
docker-compose ps

# 4. 查看日志
docker-compose logs -f
```

### 2.2 初始化数据库

```bash
# 等待MySQL和Redis启动完成(约30秒)
docker-compose exec tcp-server ./seed

# 预计耗时: 30-60分钟(取决于硬件性能)
```

### 2.3 访问应用

- HTTP Server: http://localhost:8080
- 登录页面: http://localhost:8080/web/login.html
- TCP Server: localhost:9090

### 2.4 停止服务

```bash
# 停止所有服务
docker-compose down

# 停止并删除数据卷(谨慎操作!)
docker-compose down -v
```

## 3. 本地部署

### 3.1 安装依赖

#### macOS
```bash
# 安装Go
brew install go

# 安装MySQL
brew install mysql@8.0
brew services start mysql@8.0

# 安装Redis
brew install redis
brew services start redis
```

#### Ubuntu/Debian
```bash
# 安装Go
wget https://go.dev/dl/go1.26.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.26.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# 安装MySQL
sudo apt-get install mysql-server-8.0
sudo systemctl start mysql

# 安装Redis
sudo apt-get install redis-server
sudo systemctl start redis
```

### 3.2 配置数据库

```sql
-- 创建数据库
CREATE DATABASE user_management CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 创建用户(可选)
CREATE USER 'ums'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON user_management.* TO 'ums'@'localhost';
FLUSH PRIVILEGES;
```

### 3.3 编译运行

```bash
# 1. 下载依赖
go mod download

# 2. 设置环境变量
export MYSQL_HOST=localhost
export MYSQL_PORT=3306
export MYSQL_USER=root
export MYSQL_PASSWORD=root
export MYSQL_DB=user_management
export REDIS_HOST=localhost
export REDIS_PORT=6379

# 3. 启动TCP Server
go run cmd/tcp-server/main.go &

# 4. 启动HTTP Server
go run cmd/http-server/main.go &

# 5. 插入测试数据
go run cmd/seed/main.go
```

## 4. 生产环境部署

### 4.1 系统优化(macOS)

```bash
# 增加文件描述符限制
sudo sysctl -w kern.ipc.somaxconn=2048
sudo sysctl -w kern.maxfiles=12288
ulimit -n 10000

# 持久化配置
echo "kern.ipc.somaxconn=2048" | sudo tee -a /etc/sysctl.conf
echo "kern.maxfiles=12288" | sudo tee -a /etc/sysctl.conf
```

### 4.2 系统优化(Linux)

```bash
# /etc/security/limits.conf
* soft nofile 65536
* hard nofile 65536

# /etc/sysctl.conf
net.core.somaxconn = 2048
net.ipv4.tcp_max_syn_backlog = 2048
fs.file-max = 65536

# 应用配置
sudo sysctl -p
```

### 4.3 MySQL优化

```ini
# my.cnf
[mysqld]
# 连接数
max_connections = 1000
max_connect_errors = 100000

# InnoDB
innodb_buffer_pool_size = 4G
innodb_log_file_size = 512M
innodb_flush_log_at_trx_commit = 2
innodb_flush_method = O_DIRECT

# 查询缓存
query_cache_type = 0  # MySQL 8.0已移除

# 字符集
character-set-server = utf8mb4
collation-server = utf8mb4_unicode_ci
```

### 4.4 Redis优化

```conf
# redis.conf
maxmemory 2gb
maxmemory-policy allkeys-lru

# 持久化
appendonly yes
appendfsync everysec

# 网络连接
tcp-backlog 2048
timeout 300
```

### 4.5 使用systemd管理(推荐)

#### TCP Server服务

```ini
# /etc/systemd/system/ums-tcp.service
[Unit]
Description=User Management TCP Server
After=network.target mysql.service redis.service

[Service]
Type=simple
User=ums
WorkingDirectory=/opt/ums
ExecStart=/opt/ums/tcp-server
Restart=always
RestartSec=5
Environment=MYSQL_HOST=localhost
Environment=REDIS_HOST=localhost

[Install]
WantedBy=multi-user.target
```

#### HTTP Server服务

```ini
# /etc/systemd/system/ums-http.service
[Unit]
Description=User Management HTTP Server
After=network.target ums-tcp.service

[Service]
Type=simple
User=ums
WorkingDirectory=/opt/ums
ExecStart=/opt/ums/http-server
Restart=always
RestartSec=5
Environment=TCP_PORT=tcp-server:9090

[Install]
WantedBy=multi-user.target
```

#### 启动服务

```bash
# 重载配置
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start ums-tcp
sudo systemctl start ums-http

# 开机自启
sudo systemctl enable ums-tcp
sudo systemctl enable ums-http

# 查看状态
sudo systemctl status ums-tcp
sudo systemctl status ums-http
```

## 5. 监控与告警

### 5.1 健康检查

```bash
# HTTP Server健康检查
curl http://localhost:8080/health

# MySQL健康检查
mysqladmin ping -h localhost

# Redis健康检查
redis-cli ping
```

### 5.2 性能监控

#### 使用Prometheus + Grafana

```yaml
# docker-compose.monitoring.yml
version: '3.8'
services:
  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml

  grafana:
    image: grafana/grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
```

### 5.3 日志收集

```bash
# Docker日志
docker-compose logs -f --tail=100

# 查看TCP Server日志
docker logs ums-tcp-server

# 查看HTTP Server日志
docker logs ums-http-server
```

## 6. 备份与恢复

### 6.1 MySQL备份

```bash
# 全量备份
mysqldump -u root -p \
  --single-transaction \
  --routines \
  --triggers \
  user_management > backup_$(date +%Y%m%d_%H%M%S).sql

# 压缩备份
mysqldump -u root -p user_management | gzip > backup_$(date +%Y%m%d_%H%M%S).sql.gz
```

### 6.2 MySQL恢复

```bash
# 从备份恢复
mysql -u root -p user_management < backup_20240101_120000.sql

# 从压缩备份恢复
gunzip < backup_20240101_120000.sql.gz | mysql -u root -p user_management
```

### 6.3 Redis备份

```bash
# 触发RDB快照
redis-cli BGSAVE

# 备份RDB文件
cp /var/lib/redis/dump.rdb /backup/redis_dump_$(date +%Y%m%d).rdb
```

### 6.4 自动化备份脚本

```bash
#!/bin/bash
# backup.sh

BACKUP_DIR="/backup/ums"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR

# 备份MySQL
mysqldump -u root -p'password' user_management | gzip > $BACKUP_DIR/mysql_$DATE.sql.gz

# 备份Redis
redis-cli BGSAVE
sleep 5
cp /var/lib/redis/dump.rdb $BACKUP_DIR/redis_$DATE.rdb

# 保留最近7天备份
find $BACKUP_DIR -name "*.gz" -mtime +7 -delete
find $BACKUP_DIR -name "*.rdb" -mtime +7 -delete

echo "Backup completed: $DATE"
```

```bash
# 添加到crontab,每天凌晨2点执行
0 2 * * * /opt/ums/backup.sh >> /var/log/ums_backup.log 2>&1
```

## 7. 故障排查

### 7.1 常见问题

#### 问题1: 连接MySQL失败

```bash
# 检查MySQL是否运行
docker-compose ps mysql
# 或
systemctl status mysql

# 检查端口
netstat -tlnp | grep 3306

# 查看MySQL日志
docker logs ums-mysql
# 或
tail -f /var/log/mysql/error.log
```

#### 问题2: Redis连接超时

```bash
# 检查Redis状态
docker-compose ps redis
redis-cli ping

# 检查内存使用
docker stats ums-redis
redis-cli INFO memory
```

#### 问题3: QPS不达标

```bash
# 检查系统资源
top
htop
vmstat 1

# 检查网络连接
netstat -an | grep ESTABLISHED | wc -l

# 检查数据库慢查询
mysql -u root -p -e "SHOW PROCESSLIST;"

# 运行性能测试
go run cmd/benchmark/main.go
```

### 7.2 性能调优

```bash
# 1. 调整Go GC
export GOGC=50  # 更频繁的GC,降低内存占用

# 2. 调整MySQL连接池
# 在config/config.go中修改
MaxOpenConns: 200
MaxIdleConns: 50

# 3. 调整RPC连接池
# 在cmd/http-server/main.go中修改
rpcPool := rpc.NewClientPool(cfg.TCPServer.Port, 200)
```

## 8. 升级与维护

### 8.1 应用升级

```bash
# Docker部署
docker-compose pull
docker-compose up -d

# 本地部署
git pull
go build -o tcp-server cmd/tcp-server/main.go
go build -o http-server cmd/http-server/main.go
sudo systemctl restart ums-tcp ums-http
```

### 8.2 数据库迁移

```bash
# GORM自动迁移会在启动时执行
# 如需手动迁移,可编写迁移脚本
```

### 8.3 回滚方案

```bash
# Docker回滚
docker-compose down
docker-compose up -d --force-recreate

# 数据库回滚
mysql -u root -p user_management < backup_YYYYMMDD_HHMMSS.sql
```

## 9. 安全加固

### 9.1 防火墙配置

```bash
# Ubuntu (UFW)
sudo ufw allow 8080/tcp  # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw deny 3306/tcp   # MySQL(仅内网)
sudo ufw deny 6379/tcp   # Redis(仅内网)
sudo ufw enable

# CentOS (firewalld)
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload
```

### 9.2 SSL/TLS配置

```nginx
# Nginx反向代理配置
server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /etc/ssl/certs/cert.pem;
    ssl_certificate_key /etc/ssl/private/key.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 9.3 密码策略

```sql
-- MySQL密码策略
INSTALL PLUGIN validate_password SONAME 'validate_password.so';
SET GLOBAL validate_password_policy = STRONG;
SET GLOBAL validate_password_length = 12;
```

## 10. 联系支持

如遇问题,请提供以下信息:

1. 部署方式(Docker/本地)
2. 系统版本和配置
3. 错误日志
4. 复现步骤

---

**文档版本**: 1.0  
**最后更新**: 2024-01-01
