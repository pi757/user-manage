package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter 速率限制器
type RateLimiter struct {
	clients map[string]*clientInfo
	mu      sync.Mutex
	maxReqs int
	window  time.Duration
}

type clientInfo struct {
	count   int
	resetAt time.Time
}

// NewRateLimiter 创建速率限制器
// maxReqs: 每个窗口期最大请求数
// window: 窗口期时长
func NewRateLimiter(maxReqs int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*clientInfo),
		maxReqs: maxReqs,
		window:  window,
	}

	// 启动清理协程
	go rl.cleanup()

	return rl
}

// Middleware Gin 中间件
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用 IP 作为标识
		clientIP := c.ClientIP()

		if !rl.allow(clientIP) {
			c.JSON(429, gin.H{
				"code":    429,
				"message": "too many requests, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// allow 检查是否允许请求
func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	info, exists := rl.clients[key]

	if !exists || now.After(info.resetAt) {
		// 新客户端或窗口期已过
		rl.clients[key] = &clientInfo{
			count:   1,
			resetAt: now.Add(rl.window),
		}
		return true
	}

	info.count++
	return info.count <= rl.maxReqs
}

// cleanup 定期清理过期记录
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, info := range rl.clients {
			if now.After(info.resetAt) {
				delete(rl.clients, key)
			}
		}
		rl.mu.Unlock()
	}
}
