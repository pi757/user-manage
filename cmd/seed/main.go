package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"
	"user-management-system/auth"
	"user-management-system/config"
	"user-management-system/database"
	"user-management-system/models"
)

const (
	totalUsers    = 10000000 // 1000万用户
	batchSize     = 10000    // 每批次插入数量
	concurrentNum = 10       // 并发协程数
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化数据库
	if err := database.InitMySQL(&cfg.MySQL); err != nil {
		log.Fatalf("Failed to initialize MySQL: %v", err)
	}

	// 自动迁移表结构
	if err := database.AutoMigrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	fmt.Printf("Starting to insert %d users...\n", totalUsers)
	startTime := time.Now()

	// 使用批量插入
	insertUsersBatch()

	elapsed := time.Since(startTime)
	fmt.Printf("Completed! Total time: %v\n", elapsed)
	fmt.Printf("Average speed: %.2f users/second\n", float64(totalUsers)/elapsed.Seconds())
}

func insertUsersBatch() {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrentNum)

	for i := 0; i < totalUsers; i += batchSize {
		wg.Add(1)
		semaphore <- struct{}{} // 限制并发数

		go func(offset int) {
			defer wg.Done()
			defer func() { <-semaphore }()

			batchEnd := offset + batchSize
			if batchEnd > totalUsers {
				batchEnd = totalUsers
			}

			users := generateUsers(offset+1, batchEnd)

			// 分批插入,每批1000条
			for j := 0; j < len(users); j += 1000 {
				end := j + 1000
				if end > len(users) {
					end = len(users)
				}

				if err := database.DB.Create(new(users[j:end])).Error; err != nil {
					log.Printf("Failed to insert batch %d-%d: %v", j, end, err)
					return
				}
			}

			fmt.Printf("Inserted users %d-%d\n", offset+1, batchEnd)
		}(i)
	}

	wg.Wait()
}

func generateUsers(start, end int) []models.User {
	users := make([]models.User, 0, end-start)

	for i := start; i <= end; i++ {
		password, _ := auth.HashPassword(fmt.Sprintf("password%d", i))

		user := models.User{
			UID:          fmt.Sprintf("uid_%d", i),
			Username:     fmt.Sprintf("user%d", i),
			PasswordHash: password,
			Nickname:     generateNickname(i),
			Avatar:       "",
			IsAvailable:  1,
		}
		users = append(users, user)
	}

	return users
}

func generateNickname(index int) string {
	// 生成支持Unicode的nickname
	nicknames := []string{
		"张三", "李四", "王五", "赵六", "钱七",
		"孙八", "周九", "吴十", "郑十一", "王十二",
		"Alice", "Bob", "Charlie", "David", "Eve",
		"Frank", "Grace", "Henry", "Ivy", "Jack",
		"ユーザー", "利用者", "会員", "顧客", "訪問者",
		"Пользователь", "Клиент", "Гость", "Участник", "Посетитель",
	}

	rand.Seed(time.Now().UnixNano())
	base := nicknames[rand.Intn(len(nicknames))]
	return fmt.Sprintf("%s_%d", base, index)
}

// 优化版本:使用原生SQL批量插入(更快)
func insertUsersWithRawSQL() {
	fmt.Println("Using raw SQL for better performance...")

	batchSQL := "INSERT INTO user (uid, username, password_hash, nickname, avatar, is_available, create_time, update_time) VALUES "
	valueTemplate := "(?, ?, ?, ?, ?, 1, NOW(), NOW())"

	for i := 1; i <= totalUsers; i += batchSize {
		end := i + batchSize
		if end > totalUsers {
			end = totalUsers
		}

		args := make([]interface{}, 0, batchSize*5)
		var values []string

		for j := i; j < end; j++ {
			password, _ := auth.HashPassword(fmt.Sprintf("password%d", j))
			nickname := generateNickname(j)

			values = append(values, valueTemplate)
			args = append(args, fmt.Sprintf("uid_%d", j), fmt.Sprintf("user%d", j), password, nickname, "")
		}

		sql := batchSQL + strings.Join(values, ",")

		if err := database.DB.Exec(sql, args...).Error; err != nil {
			log.Printf("Failed to insert batch %d-%d: %v", i, end, err)
			continue
		}

		fmt.Printf("Inserted users %d-%d\n", i, end-1)
	}
}
