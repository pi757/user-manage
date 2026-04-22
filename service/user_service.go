package service

import (
	"fmt"
	"strconv"
	"user-management-system/auth"
	"user-management-system/database"
	"user-management-system/models"
)

// UserService 用户服务
type UserService struct {
	authService *auth.AuthService
}

// NewUserService 创建用户服务
func NewUserService(authService *auth.AuthService) *UserService {
	return &UserService{
		authService: authService,
	}
}

// Login 用户登录
func (s *UserService) Login(params map[string]interface{}) (interface{}, error) {
	username, ok := params["username"].(string)
	if !ok || username == "" {
		return nil, fmt.Errorf("username is required")
	}

	password, ok := params["password"].(string)
	if !ok || password == "" {
		return nil, fmt.Errorf("password is required")
	}

	// 查询用户
	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	// 验证密码
	if !auth.CheckPassword(password, user.Password) {
		return nil, fmt.Errorf("invalid username or password")
	}

	// 创建session
	token, err := s.authService.CreateSession(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return map[string]interface{}{
		"token":    token,
		"user_id":  user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
	}, nil
}

// GetProfile 获取用户信息
func (s *UserService) GetProfile(params map[string]interface{}) (interface{}, error) {
	token, ok := params["token"].(string)
	if !ok || token == "" {
		return nil, fmt.Errorf("token is required")
	}

	// 验证session
	userID, err := s.authService.ValidateSession(token)
	if err != nil {
		return nil, fmt.Errorf("invalid session: %w", err)
	}

	// 查询用户信息
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
	}, nil
}

// UpdateProfile 更新用户信息
func (s *UserService) UpdateProfile(params map[string]interface{}) (interface{}, error) {
	token, ok := params["token"].(string)
	if !ok || token == "" {
		return nil, fmt.Errorf("token is required")
	}

	// 验证session
	userID, err := s.authService.ValidateSession(token)
	if err != nil {
		return nil, fmt.Errorf("invalid session: %w", err)
	}

	// 查询用户
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// 更新nickname(支持Unicode)
	if nickname, ok := params["nickname"].(string); ok && nickname != "" {
		user.Nickname = nickname
	}

	// 更新avatar
	if avatar, ok := params["avatar"].(string); ok && avatar != "" {
		user.Avatar = avatar
	}

	// 保存更新
	if err := database.DB.Save(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
	}, nil
}

// Logout 用户登出
func (s *UserService) Logout(params map[string]interface{}) (interface{}, error) {
	token, ok := params["token"].(string)
	if !ok || token == "" {
		return nil, fmt.Errorf("token is required")
	}

	if err := s.authService.DeleteSession(token); err != nil {
		return nil, fmt.Errorf("failed to logout: %w", err)
	}

	return map[string]interface{}{
		"message": "logout successful",
	}, nil
}

// GetUserByID 根据ID获取用户(用于测试)
func (s *UserService) GetUserByID(params map[string]interface{}) (interface{}, error) {
	idFloat, ok := params["id"].(float64)
	if !ok {
		return nil, fmt.Errorf("id is required")
	}
	id := uint(idFloat)

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
	}, nil
}

// BatchGetUsers 批量获取用户(用于性能测试)
func (s *UserService) BatchGetUsers(params map[string]interface{}) (interface{}, error) {
	idsParam, ok := params["ids"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("ids is required")
	}

	var ids []uint
	for _, idFloat := range idsParam {
		if id, ok := idFloat.(float64); ok {
			ids = append(ids, uint(id))
		}
	}

	var users []models.User
	if err := database.DB.Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	result := make([]map[string]interface{}, 0, len(users))
	for _, user := range users {
		result = append(result, map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"nickname": user.Nickname,
			"avatar":   user.Avatar,
		})
	}

	return result, nil
}

// ValidateToken 验证token(用于性能测试)
func (s *UserService) ValidateToken(params map[string]interface{}) (interface{}, error) {
	token, ok := params["token"].(string)
	if !ok || token == "" {
		return nil, fmt.Errorf("token is required")
	}

	userID, err := s.authService.ValidateSession(token)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"valid":    true,
		"user_id":  userID,
	}, nil
}

// ParseUint 辅助函数:将float64转换为uint
func ParseUint(value interface{}) (uint, error) {
	switch v := value.(type) {
	case float64:
		return uint(v), nil
	case string:
		id, err := strconv.ParseUint(v, 10, 32)
		return uint(id), err
	default:
		return 0, fmt.Errorf("invalid type for uint conversion")
	}
}
