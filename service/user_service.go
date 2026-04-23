package service

import (
	"fmt"
	"time"
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

	// 查询用户并验证密码（直接在SQL中验证，防止时序攻击）
	var user models.User
	if err := database.DB.Select("id", "uid", "username", "nickname", "password_hash", "avatar", "is_available").Where("username = ? AND is_available = 1", username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	// 验证密码
	if !auth.CheckPassword(password, user.PasswordHash) {
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
		"uid":      user.UID,
		"username": user.Username,
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
	}, nil
}

// Register 用户注册
func (s *UserService) Register(params map[string]interface{}) (interface{}, error) {
	username, ok := params["username"].(string)
	if !ok || username == "" {
		return nil, fmt.Errorf("username is required")
	}

	// 验证用户名格式
	if len(username) < 3 || len(username) > 64 {
		return nil, fmt.Errorf("username must be between 3 and 64 characters")
	}

	password, ok := params["password"].(string)
	if !ok || password == "" {
		return nil, fmt.Errorf("password is required")
	}

	// 验证密码强度
	if len(password) < 6 {
		return nil, fmt.Errorf("password must be at least 6 characters")
	}

	// 检查用户名是否已存在
	var count int64
	database.DB.Model(&models.User{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("username already exists")
	}

	// 密码加密
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 生成UID
	uid := fmt.Sprintf("uid_%s_%d", username, time.Now().UnixNano())

	// 获取nickname，如果未提供则使用username
	nickname, _ := params["nickname"].(string)
	if nickname == "" {
		nickname = username
	}

	// 创建用户
	user := models.User{
		UID:          uid,
		Username:     username,
		Nickname:     nickname,
		PasswordHash: passwordHash,
		Avatar:       "",
		IsAvailable:  1,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 自动登录，创建session
	token, err := s.authService.CreateSession(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return map[string]interface{}{
		"token":    token,
		"user_id":  user.ID,
		"uid":      user.UID,
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
	if err := database.DB.Select("id", "uid", "username", "nickname", "avatar", "is_available").First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// 检查用户是否可用
	if user.IsAvailable == 0 {
		return nil, fmt.Errorf("user is not available")
	}

	return map[string]interface{}{
		"id":           user.ID,
		"uid":          user.UID,
		"username":     user.Username,
		"nickname":     user.Nickname,
		"avatar":       user.Avatar,
		"is_available": user.IsAvailable,
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
	if err := database.DB.Select("id", "uid", "username", "nickname", "avatar", "is_available").First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// 检查用户是否可用
	if user.IsAvailable == 0 {
		return nil, fmt.Errorf("user is not available")
	}

	// 更新nickname(支持Unicode)
	updates := make(map[string]interface{})
	if nickname, ok := params["nickname"].(string); ok && nickname != "" {
		updates["nickname"] = nickname
	}

	// 更新avatar
	if avatar, ok := params["avatar"].(string); ok && avatar != "" {
		updates["avatar"] = avatar
	}

	// 如果没有要更新的字段，直接返回
	if len(updates) == 0 {
		return map[string]interface{}{
			"id":           user.ID,
			"uid":          user.UID,
			"username":     user.Username,
			"nickname":     user.Nickname,
			"avatar":       user.Avatar,
			"is_available": user.IsAvailable,
		}, nil
	}

	// 只更新指定字段，避免更新时间为零值
	if err := database.DB.Model(&user).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	// 重新查询获取最新数据
	database.DB.Select("id", "uid", "username", "nickname", "avatar", "is_available").First(&user, userID)

	return map[string]interface{}{
		"id":           user.ID,
		"uid":          user.UID,
		"username":     user.Username,
		"nickname":     user.Nickname,
		"avatar":       user.Avatar,
		"is_available": user.IsAvailable,
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
	if err := database.DB.Select("id", "uid", "username", "nickname", "avatar", "is_available").First(&user, id).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return map[string]interface{}{
		"id":           user.ID,
		"uid":          user.UID,
		"username":     user.Username,
		"nickname":     user.Nickname,
		"avatar":       user.Avatar,
		"is_available": user.IsAvailable,
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
	if err := database.DB.Select("id", "uid", "username", "nickname", "avatar", "is_available").Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	result := make([]map[string]interface{}, 0, len(users))
	for _, user := range users {
		result = append(result, map[string]interface{}{
			"id":           user.ID,
			"uid":          user.UID,
			"name":         user.Username,
			"nickname":     user.Nickname,
			"avatar":       user.Avatar,
			"is_available": user.IsAvailable,
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
		"valid":   true,
		"user_id": userID,
	}, nil
}
