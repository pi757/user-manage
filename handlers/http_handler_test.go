package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestLoginHandler(t *testing.T) {
	// 注意: 这个测试需要mock RPC客户端
	// 这里仅作为示例
	t.Skip("Requires RPC client mock")
}

func TestResponseFormat(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	Success(c, map[string]string{"key": "value"})
	
	var response Response
	json.Unmarshal(w.Body.Bytes(), &response)
	
	if response.Code != 0 {
		t.Errorf("Expected code 0, got %d", response.Code)
	}
	
	if response.Message != "success" {
		t.Errorf("Expected message 'success', got '%s'", response.Message)
	}
}

func TestErrorResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	Error(c, http.StatusBadRequest, "test error")
	
	var response Response
	json.Unmarshal(w.Body.Bytes(), &response)
	
	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected code %d, got %d", http.StatusBadRequest, response.Code)
	}
	
	if response.Message != "test error" {
		t.Errorf("Expected message 'test error', got '%s'", response.Message)
	}
}
