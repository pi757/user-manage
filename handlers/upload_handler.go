package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"user-management-system/rpc"

	"github.com/gin-gonic/gin"
)

// UploadHandler 文件上传处理器
type UploadHandler struct {
	rpcPool     *rpc.ClientPool
	uploadDir   string
	maxFileSize int64
}

// NewUploadHandler 创建文件上传处理器
func NewUploadHandler(rpcPool *rpc.ClientPool, uploadDir string, maxFileSize int64) *UploadHandler {
	// 确保上传目录存在
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create upload directory: %v", err))
	}

	return &UploadHandler{
		rpcPool:     rpcPool,
		uploadDir:   uploadDir,
		maxFileSize: maxFileSize,
	}
}

// UploadAvatar 上传头像
func (h *UploadHandler) UploadAvatar(c *gin.Context) {
	token, err := extractToken(c)
	if err != nil {
		Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	// 验证token
	params := map[string]interface{}{
		"token": token,
	}
	resp, err := h.rpcPool.CallWithPool("user.validateToken", params)
	if err != nil || resp.Error != "" {
		Error(c, http.StatusUnauthorized, "invalid token")
		return
	}

	// 获取用户ID和当前头像
	resultMap := resp.Result.(map[string]interface{})
	userID := convertToUint(resultMap["user_id"])

	// 获取用户当前头像（用于删除旧文件）
	profileParams := map[string]interface{}{
		"token": token,
	}
	profileResp, err := h.rpcPool.CallWithPool("user.getProfile", profileParams)
	var oldAvatar string
	if err == nil && profileResp.Error == "" {
		if profileResult, ok := profileResp.Result.(map[string]interface{}); ok {
			if avatar, ok := profileResult["avatar"].(string); ok && avatar != "" {
				oldAvatar = avatar
			}
		}
	}

	// 处理文件上传
	file, err := c.FormFile("avatar")
	if err != nil {
		Error(c, http.StatusBadRequest, "failed to get file: "+err.Error())
		return
	}

	// 验证文件大小
	if file.Size > h.maxFileSize {
		Error(c, http.StatusBadRequest, fmt.Sprintf("file size exceeds limit (%d bytes)", h.maxFileSize))
		return
	}

	// 验证文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
	}
	if !allowedExts[ext] {
		Error(c, http.StatusBadRequest, "invalid file type, only jpg, jpeg, png, gif are allowed")
		return
	}

	// 生成唯一文件名
	filename := fmt.Sprintf("%d_%d%s", userID, time.Now().UnixNano(), ext)
	filePath := filepath.Join(h.uploadDir, filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		Error(c, http.StatusInternalServerError, "failed to save file: "+err.Error())
		return
	}

	// 更新用户头像URL
	avatarURL := fmt.Sprintf("/uploads/%s", filename)
	updateParams := map[string]interface{}{
		"token":  token,
		"avatar": avatarURL,
	}

	updateResp, err := h.rpcPool.CallWithPool("user.updateProfile", updateParams)
	if err != nil || updateResp.Error != "" {
		// 删除已上传的文件
		err := os.Remove(filePath)
		if err != nil {
			return
		}
		Error(c, http.StatusInternalServerError, "failed to update avatar")
		return
	}

	// 删除旧头像文件
	if oldAvatar != "" {
		// 提取文件名（去除 /uploads/ 前缀）
		oldFilename := strings.TrimPrefix(oldAvatar, "/uploads/")
		oldFilepath := filepath.Join(h.uploadDir, oldFilename)
		if err := os.Remove(oldFilepath); err == nil {
			fmt.Printf("Deleted old avatar: %s\n", oldFilepath)
		} else {
			fmt.Printf("Warning: failed to delete old avatar %s: %v\n", oldFilepath, err)
		}
	}

	Success(c, map[string]interface{}{
		"avatar_url": avatarURL,
	})
}

// convertToUint 将interface{}转换为uint，兼容JSON和MessagePack的不同数字类型
func convertToUint(v interface{}) uint {
	switch val := v.(type) {
	case float64:
		return uint(val)
	case int:
		return uint(val)
	case int8:
		return uint(val)
	case int16:
		return uint(val)
	case int32:
		return uint(val)
	case int64:
		return uint(val)
	case uint:
		return val
	case uint8:
		return uint(val)
	case uint16:
		return uint(val)
	case uint32:
		return uint(val)
	case uint64:
		return uint(val)
	default:
		return 0
	}
}
