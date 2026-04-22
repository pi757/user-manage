package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
	"user-management-system/config"
	"user-management-system/database"
	"user-management-system/models"

	"golang.org/x/crypto/bcrypt"
)

// AuthService 认证服务
type AuthService struct {
	config *config.SessionConfig
}

// NewAuthService 创建认证服务
func NewAuthService(cfg *config.SessionConfig) *AuthService {
	return &AuthService{
		config: cfg,
	}
}

// GenerateToken 生成随机token
func (s *AuthService) GenerateToken() (string, error) {
	bytes := make([]byte, s.config.TokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// HashPassword 密码哈希
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// CheckPassword 验证密码
func CheckPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// CreateSession 创建会话
func (s *AuthService) CreateSession(userID uint) (string, error) {
	token, err := s.GenerateToken()
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(time.Duration(s.config.Expiration) * time.Second)

	// 存储到Redis
	ctx := context.Background()
	key := fmt.Sprintf("session:%s", token)
	err = database.RedisClient.Set(ctx, key, userID, time.Duration(s.config.Expiration)*time.Second).Err()
	if err != nil {
		return "", fmt.Errorf("failed to store session in Redis: %w", err)
	}

	// 同时存储到MySQL(用于持久化)
	session := models.Session{
		Token:     token,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}
	if err := database.DB.Create(&session).Error; err != nil {
		return "", fmt.Errorf("failed to store session in MySQL: %w", err)
	}

	return token, nil
}

// ValidateSession 验证会话
func (s *AuthService) ValidateSession(token string) (uint, error) {
	ctx := context.Background()
	key := fmt.Sprintf("session:%s", token)

	// 先从Redis查询
	userIDStr, err := database.RedisClient.Get(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("invalid or expired session")
	}

	var userID uint
	fmt.Sscanf(userIDStr, "%d", &userID)

	// 刷新过期时间
	database.RedisClient.Expire(ctx, key, time.Duration(s.config.Expiration)*time.Second)

	return userID, nil
}

// DeleteSession 删除会话
func (s *AuthService) DeleteSession(token string) error {
	ctx := context.Background()
	key := fmt.Sprintf("session:%s", token)

	// 从Redis删除
	if err := database.RedisClient.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete session from Redis: %w", err)
	}

	// 从MySQL删除
	if err := database.DB.Where("token = ?", token).Delete(&models.Session{}).Error; err != nil {
		return fmt.Errorf("failed to delete session from MySQL: %w", err)
	}

	return nil
}
