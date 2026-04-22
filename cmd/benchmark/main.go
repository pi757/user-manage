package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const (
	baseURL       = "http://localhost:8080"
	loginURL      = baseURL + "/api/login"
	profileURL    = baseURL + "/api/profile"
	logoutURL     = baseURL + "/api/logout"
	testUserCount = 200 // 测试用户数量
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Token    string `json:"token"`
		UserID   uint   `json:"user_id"`
		Username string `json:"username"`
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
	} `json:"data"`
}

type ProfileResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
	} `json:"data"`
}

// TestResult 测试结果
type TestResult struct {
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64
	TotalTime       time.Duration
	AvgLatency      time.Duration
	QPS             float64
}

func main() {
	fmt.Println("=== User Management System Performance Test ===\n")

	// 测试场景1: 200并发固定用户
	fmt.Println("Test 1: 200 concurrent users (fixed)")
	result1 := runTest(200, 10000, false)
	printResult(result1)

	// 等待一下
	time.Sleep(5 * time.Second)

	// 测试场景2: 200并发随机用户
	fmt.Println("\nTest 2: 200 concurrent users (random)")
	result2 := runTest(200, 10000, true)
	printResult(result2)

	// 等待一下
	time.Sleep(5 * time.Second)

	// 测试场景3: 2000并发固定用户
	fmt.Println("\nTest 3: 2000 concurrent users (fixed)")
	result3 := runTest(2000, 10000, false)
	printResult(result3)

	// 等待一下
	time.Sleep(5 * time.Second)

	// 测试场景4: 2000并发随机用户
	fmt.Println("\nTest 4: 2000 concurrent users (random)")
	result4 := runTest(2000, 10000, true)
	printResult(result4)
}

func runTest(concurrency, totalRequests int, randomUser bool) TestResult {
	fmt.Printf("Starting test: concurrency=%d, totalRequests=%d, randomUser=%v\n",
		concurrency, totalRequests, randomUser)

	// 预登录获取tokens
	tokens := make([]string, testUserCount)
	var wg sync.WaitGroup

	for i := 0; i < testUserCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			username := fmt.Sprintf("user%d", idx+1)
			password := fmt.Sprintf("password%d", idx+1)

			token, err := login(username, password)
			if err != nil {
				fmt.Printf("Failed to login user %s: %v\n", username, err)
				return
			}
			tokens[idx] = token
		}(i)
	}
	wg.Wait()

	// 开始性能测试
	var totalReqs, successReqs, failedReqs int64
	var totalLatency int64
	var latencyMu sync.Mutex

	startTime := time.Now()
	requestsPerWorker := totalRequests / concurrency

	var workerWg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		workerWg.Add(1)
		go func(workerID int) {
			defer workerWg.Done()

			client := &http.Client{
				Timeout: 30 * time.Second,
			}

			for j := 0; j < requestsPerWorker; j++ {
				// 选择token
				var token string
				if randomUser {
					token = tokens[rand.Intn(len(tokens))]
				} else {
					token = tokens[workerID%len(tokens)]
				}

				// 发送请求
				reqStart := time.Now()
				success := getProfile(client, token)
				reqDuration := time.Since(reqStart)

				atomic.AddInt64(&totalReqs, 1)
				if success {
					atomic.AddInt64(&successReqs, 1)
				} else {
					atomic.AddInt64(&failedReqs, 1)
				}

				latencyMu.Lock()
				totalLatency += int64(reqDuration)
				latencyMu.Unlock()
			}
		}(i)
	}

	workerWg.Wait()
	totalTime := time.Since(startTime)

	return TestResult{
		TotalRequests:   atomic.LoadInt64(&totalReqs),
		SuccessRequests: atomic.LoadInt64(&successReqs),
		FailedRequests:  atomic.LoadInt64(&failedReqs),
		TotalTime:       totalTime,
		AvgLatency:      time.Duration(totalLatency / totalReqs),
		QPS:             float64(totalReqs) / totalTime.Seconds(),
	}
}

func login(username, password string) (string, error) {
	reqBody := LoginRequest{
		Username: username,
		Password: password,
	}

	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post(loginURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var loginResp LoginResponse
	json.Unmarshal(body, &loginResp)

	if loginResp.Code != 0 {
		return "", fmt.Errorf("login failed: %s", loginResp.Message)
	}

	return loginResp.Data.Token, nil
}

func getProfile(client *http.Client, token string) bool {
	req, _ := http.NewRequest("GET", profileURL, nil)
	req.Header.Set("Authorization", token)

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	body, _ := io.ReadAll(resp.Body)
	var profileResp ProfileResponse
	json.Unmarshal(body, &profileResp)

	return profileResp.Code == 0
}

func printResult(result TestResult) {
	fmt.Println("\n--- Test Results ---")
	fmt.Printf("Total Requests:    %d\n", result.TotalRequests)
	fmt.Printf("Success Requests:  %d\n", result.SuccessRequests)
	fmt.Printf("Failed Requests:   %d\n", result.FailedRequests)
	fmt.Printf("Total Time:        %v\n", result.TotalTime)
	fmt.Printf("Average Latency:   %v\n", result.AvgLatency)
	fmt.Printf("QPS:               %.2f\n", result.QPS)

	// 检查是否达标
	if result.QPS >= 3000 {
		fmt.Println("✓ PASSED (QPS >= 3000)")
	} else if result.QPS >= 1000 {
		fmt.Println("⚠ PARTIAL (1000 <= QPS < 3000)")
	} else {
		fmt.Println("✗ FAILED (QPS < 1000)")
	}
}
