package config

// TCPServiceConfig TCP服务配置
type TCPServiceConfig struct {
	TCPServer TCPServerConfig
	MySQL     MySQLConfig
	Redis     RedisConfig
	Session   SessionConfig
}

// LoadTCPServiceConfig 加载TCP服务配置
func LoadTCPServiceConfig() *TCPServiceConfig {
	return &TCPServiceConfig{
		TCPServer: TCPServerConfig{
			Port: getEnv("TCP_PORT", ":9090"),
		},
		MySQL: MySQLConfig{
			Host:         getEnv("MYSQL_HOST", "localhost"),
			Port:         getEnv("MYSQL_PORT", "3307"),
			User:         getEnv("MYSQL_USER", "root"),
			Password:     getEnv("MYSQL_PASSWORD", "root123"),
			DBName:       getEnv("MYSQL_DB", "user_management"),
			MaxOpenConns: 100,
			MaxIdleConns: 20,
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6380"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
			PoolSize: 100,
		},
		Session: SessionConfig{
			TokenLength: 32,
			Expiration:  3600, // 1 hour
		},
	}
}
