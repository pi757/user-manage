package auth

import (
	"testing"
	"user-management-system/config"
)

// 测试密码哈希
func TestHashPassword(t *testing.T) {
	password := "test_password_123"

	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hashed == "" {
		t.Error("Hashed password should not be empty")
	}

	if hashed == password {
		t.Error("Hashed password should be different from original")
	}
	t.Logf("original:%s, Hashed password: %s", password, hashed)
}

// 校验密码
func TestCheckPassword(t *testing.T) {
	password := "test_password_123"

	hashed, _ := HashPassword(password)

	// 测试正确密码
	if !CheckPassword(password, hashed) {
		t.Error("CheckPassword should return true for correct password")
	}

	// 测试错误密码
	if CheckPassword("wrong_password", hashed) {
		t.Error("CheckPassword should return false for wrong password")
	}

	// 测试空密码
	if CheckPassword("", hashed) {
		t.Error("CheckPassword should return false for empty password")
	}
}

func TestGenerateToken(t *testing.T) {
	authService := NewAuthService(&config.SessionConfig{
		TokenLength: 32,
	})

	token1, err := authService.GenerateToken()
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	token2, err := authService.GenerateToken()
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// 验证token长度(64个hex字符)
	if len(token1) != 64 {
		t.Errorf("Token length should be 64, got %d", len(token1))
	}

	// 验证token唯一性
	if token1 == token2 {
		t.Error("Tokens should be unique")
	}
}

func TestValidateSession(t *testing.T) {
	// 注意: 这个测试需要真实的Redis连接
	// 在实际项目中应该使用mock
	t.Skip("Requires Redis connection")
}
