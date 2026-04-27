package config

// HTTPServiceConfig HTTP服务配置
type HTTPServiceConfig struct {
	HTTPServer HTTPServerConfig
	FileUpload FileUploadConfig
	Redis      RedisConfig
}

// LoadHTTPServiceConfig 加载HTTP服务配置
func LoadHTTPServiceConfig() *HTTPServiceConfig {
	return &HTTPServiceConfig{
		HTTPServer: HTTPServerConfig{
			Port: getEnv("HTTP_PORT", ":8080"),
		},
		FileUpload: FileUploadConfig{
			UploadDir: getEnv("UPLOAD_DIR", "./uploads"),
			MaxSize:   10 << 20, // 10MB
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6380"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
			PoolSize: 100,
		},
	}
}
