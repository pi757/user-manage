package main

import (
	"fmt"
)

// 用户管理系统 - User Management System
// 
// 这是一个高性能的用户管理系统,采用自研RPC框架和微服务架构
//
// 快速开始:
// 1. Docker部署(推荐):
//    docker-compose up -d
//    docker-compose exec tcp-server ./seed
//
// 2. 本地运行:
//    go run cmd/tcp-server/main.go   # 终端1
//    go run cmd/http-server/main.go  # 终端2
//    go run cmd/seed/main.go         # 终端3 (插入测试数据)
//
// 访问应用:
// - HTTP Server: http://localhost:8080
// - 登录页面: http://localhost:8080/web/login.html
//
// 性能测试:
//    go run cmd/benchmark/main.go
//
// 详细文档请查看 docs/ 目录
func main() {
	fmt.Println("用户管理系统 (User Management System)")
	fmt.Println("")
	fmt.Println("请选择运行的组件:")
	fmt.Println("  1. TCP Server:  go run cmd/tcp-server/main.go")
	fmt.Println("  2. HTTP Server: go run cmd/http-server/main.go")
	fmt.Println("  3. Seed Data:   go run cmd/seed/main.go")
	fmt.Println("  4. Benchmark:   go run cmd/benchmark/main.go")
	fmt.Println("")
	fmt.Println("或使用Docker一键部署:")
	fmt.Println("  docker-compose up -d")
	fmt.Println("")
	fmt.Println("更多信息请查看 README.md")
}
