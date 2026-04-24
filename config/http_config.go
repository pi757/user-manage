package config

// HTTPServiceConfig HTTP服务配置
type HTTPServiceConfig struct {
	HTTPServer HTTPServerConfig
	TCPServer  TCPServerConfig
	FileUpload FileUploadConfig
}

// LoadHTTPServiceConfig 加载HTTP服务配置
func LoadHTTPServiceConfig() *HTTPServiceConfig {
	return &HTTPServiceConfig{
		HTTPServer: HTTPServerConfig{
			Port: getEnv("HTTP_PORT", ":8080"),
		},
		TCPServer: TCPServerConfig{
			Port: getEnv("TCP_PORT", ":9090"),
		},
		FileUpload: FileUploadConfig{
			UploadDir: getEnv("UPLOAD_DIR", "./uploads"),
			MaxSize:   10 << 20, // 10MB
		},
	}
}
