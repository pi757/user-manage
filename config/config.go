package config

import (
	"os"
)

// HTTPServerConfig HTTP服务器配置
type HTTPServerConfig struct {
	Port string
}

// TCPServerConfig TCP服务器配置
type TCPServerConfig struct {
	Port string
}

// MySQLConfig MySQL数据库配置
type MySQLConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	DBName       string
	MaxOpenConns int
	MaxIdleConns int
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	PoolSize int
}

// SessionConfig Session配置
type SessionConfig struct {
	TokenLength int
	Expiration  int // seconds
}

// FileUploadConfig 文件上传配置
type FileUploadConfig struct {
	UploadDir string
	MaxSize   int64 // bytes
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
