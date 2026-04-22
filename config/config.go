package config

import (
	"os"
)

// Config 应用配置
type Config struct {
	HTTPServer   HTTPServerConfig
	TCPServer    TCPServerConfig
	MySQL        MySQLConfig
	Redis        RedisConfig
	Session      SessionConfig
	FileUpload   FileUploadConfig
}

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
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
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

// LoadConfig 加载配置
func LoadConfig() *Config {
	return &Config{
		HTTPServer: HTTPServerConfig{
			Port: getEnv("HTTP_PORT", ":8080"),
		},
		TCPServer: TCPServerConfig{
			Port: getEnv("TCP_PORT", ":9090"),
		},
		MySQL: MySQLConfig{
			Host:         getEnv("MYSQL_HOST", "localhost"),
			Port:         getEnv("MYSQL_PORT", "3306"),
			User:         getEnv("MYSQL_USER", "root"),
			Password:     getEnv("MYSQL_PASSWORD", "root"),
			DBName:       getEnv("MYSQL_DB", "user_management"),
			MaxOpenConns: 100,
			MaxIdleConns: 20,
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
			PoolSize: 100,
		},
		Session: SessionConfig{
			TokenLength: 32,
			Expiration:  3600, // 1 hour
		},
		FileUpload: FileUploadConfig{
			UploadDir: getEnv("UPLOAD_DIR", "./uploads"),
			MaxSize:   10 << 20, // 10MB
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
