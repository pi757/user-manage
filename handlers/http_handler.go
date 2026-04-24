package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"user-management-system/rpc"

	"github.com/gin-gonic/gin"
)

// extractToken 从 Authorization header 中提取 Token
func extractToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is required")
	}

	// 支持 Bearer Token 格式
	if strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			return "", fmt.Errorf("invalid token format")
		}
		return token, nil
	}

	// 兼容直接传递 Token 的格式（向后兼容）
	return authHeader, nil
}

// Response 统一响应格式
type Handler struct {
	rpcPool *rpc.ClientPool
}

// NewHandler 创建HTTP处理器
func NewHandler(rpcPool *rpc.ClientPool) *Handler {
	return &Handler{
		rpcPool: rpcPool,
	}
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
	})
}

// Login 登录接口
func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	params := map[string]interface{}{
		"username": req.Username,
		"password": req.Password,
	}

	resp, err := h.rpcPool.CallWithPool("user.login", params)
	if err != nil {
		Error(c, http.StatusInternalServerError, "login failed: "+err.Error())
		return
	}

	if resp.Error != "" {
		Error(c, http.StatusUnauthorized, resp.Error)
		return
	}

	Success(c, resp.Result)
}

// Register 注册接口
func (h *Handler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,min=3,max=64"`
		Password string `json:"password" binding:"required,min=6"`
		Nickname string `json:"nickname"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	params := map[string]interface{}{
		"username": req.Username,
		"password": req.Password,
		"nickname": req.Nickname,
	}

	resp, err := h.rpcPool.CallWithPool("user.register", params)
	if err != nil {
		Error(c, http.StatusInternalServerError, "register failed: "+err.Error())
		return
	}

	if resp.Error != "" {
		Error(c, http.StatusBadRequest, resp.Error)
		return
	}

	Success(c, resp.Result)
}

// GetProfile 获取用户信息
func (h *Handler) GetProfile(c *gin.Context) {
	token, err := extractToken(c)
	if err != nil {
		Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	params := map[string]interface{}{
		"token": token,
	}

	resp, err := h.rpcPool.CallWithPool("user.getProfile", params)
	if err != nil {
		Error(c, http.StatusInternalServerError, "failed to get profile: "+err.Error())
		return
	}

	if resp.Error != "" {
		Error(c, http.StatusUnauthorized, resp.Error)
		return
	}

	Success(c, resp.Result)
}

// UpdateProfile 更新用户信息
func (h *Handler) UpdateProfile(c *gin.Context) {
	token, err := extractToken(c)
	if err != nil {
		Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	var req struct {
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	params := map[string]interface{}{
		"token":    token,
		"nickname": req.Nickname,
		"avatar":   req.Avatar,
	}

	resp, err := h.rpcPool.CallWithPool("user.updateProfile", params)
	if err != nil {
		Error(c, http.StatusInternalServerError, "failed to update profile: "+err.Error())
		return
	}

	if resp.Error != "" {
		Error(c, http.StatusInternalServerError, resp.Error)
		return
	}

	Success(c, resp.Result)
}

// Logout 登出接口
func (h *Handler) Logout(c *gin.Context) {
	token, err := extractToken(c)
	if err != nil {
		Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	params := map[string]interface{}{
		"token": token,
	}

	resp, err := h.rpcPool.CallWithPool("user.logout", params)
	if err != nil {
		Error(c, http.StatusInternalServerError, "logout failed: "+err.Error())
		return
	}

	if resp.Error != "" {
		Error(c, http.StatusInternalServerError, resp.Error)
		return
	}

	Success(c, resp.Result)
}
