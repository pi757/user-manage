#!/bin/bash

# 用户管理系统 - 快速启动脚本

set -e

echo "======================================"
echo "  用户管理系统 - 快速启动"
echo "======================================"
echo ""

# 检查Docker是否安装
if ! command -v docker &> /dev/null; then
    echo "❌ Docker未安装,请先安装Docker"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose未安装,请先安装Docker Compose"
    exit 1
fi

echo "✅ Docker环境检查通过"
echo ""

# 询问部署方式
echo "请选择部署方式:"
echo "1. Docker部署 (推荐)"
echo "2. 本地运行"
echo "3. 仅启动基础设施(MySQL + Redis)"
echo ""
read -p "请输入选项 (1/2/3): " choice

case $choice in
    1)
        echo ""
        echo "🚀 开始Docker部署..."
        echo ""
        
        # 停止旧容器
        echo "📦 停止旧容器..."
        docker-compose down 2>/dev/null || true
        
        # 启动服务
        echo "📦 启动所有服务..."
        docker-compose up -d
        
        echo ""
        echo "⏳ 等待服务启动(约30秒)..."
        sleep 30
        
        # 检查服务状态
        echo ""
        echo "🔍 检查服务状态..."
        docker-compose ps
        
        echo ""
        echo "✅ 服务启动完成!"
        echo ""
        echo "📊 下一步: 插入测试数据"
        read -p "是否现在插入1000万测试数据? (这可能需要30-60分钟) [y/N]: " seed_choice
        
        if [[ $seed_choice =~ ^[Yy]$ ]]; then
            echo ""
            echo "🌱 开始插入测试数据..."
            docker-compose exec tcp-server ./seed
        fi
        
        echo ""
        echo "======================================"
        echo "  🎉 部署完成!"
        echo "======================================"
        echo ""
        echo "📱 访问应用:"
        echo "   HTTP Server: http://localhost:8080"
        echo "   登录页面:    http://localhost:8080/web/login.html"
        echo ""
        echo "🔧 管理命令:"
        echo "   查看日志:     docker-compose logs -f"
        echo "   停止服务:     docker-compose down"
        echo "   重启服务:     docker-compose restart"
        echo ""
        ;;
        
    2)
        echo ""
        echo "🚀 开始本地部署..."
        echo ""
        
        # 检查Go是否安装
        if ! command -v go &> /dev/null; then
            echo "❌ Go未安装,请先安装Go 1.26+"
            exit 1
        fi
        
        echo "✅ Go版本: $(go version)"
        echo ""
        
        # 下载依赖
        echo "📦 下载依赖..."
        go mod download
        
        echo ""
        echo "⚠️  请确保MySQL和Redis已启动"
        echo ""
        read -p "按回车键继续..." 
        
        # 启动TCP Server
        echo ""
        echo "🔧 启动TCP Server..."
        go run cmd/tcp-server/main.go &
        TCP_PID=$!
        echo "   PID: $TCP_PID"
        
        sleep 3
        
        # 启动HTTP Server
        echo "🔧 启动HTTP Server..."
        go run cmd/http-server/main.go &
        HTTP_PID=$!
        echo "   PID: $HTTP_PID"
        
        echo ""
        echo "======================================"
        echo "  🎉 服务已启动!"
        echo "======================================"
        echo ""
        echo "📱 访问应用:"
        echo "   HTTP Server: http://localhost:8080"
        echo "   登录页面:    http://localhost:8080/web/login.html"
        echo ""
        echo "🛑 停止服务: kill $TCP_PID $HTTP_PID"
        echo ""
        
        # 等待用户中断
        wait
        ;;
        
    3)
        echo ""
        echo "🚀 仅启动基础设施..."
        echo ""
        
        docker-compose up -d mysql redis
        
        echo ""
        echo "⏳ 等待服务启动..."
        sleep 10
        
        echo ""
        echo "✅ MySQL和Redis已启动"
        echo ""
        echo "🔧 连接信息:"
        echo "   MySQL: localhost:3306 (root/root)"
        echo "   Redis: localhost:6379"
        echo ""
        ;;
        
    *)
        echo "❌ 无效选项"
        exit 1
        ;;
esac
