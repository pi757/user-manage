package main

import (
	"fmt"
	"log"
	"user-management-system/auth"
	"user-management-system/config"
	"user-management-system/database"
	"user-management-system/models"
)

const testUsers = 200 // 仅创建200个测试用户

func main() {
	cfg := config.LoadConfig()

	if err := database.InitMySQL(&cfg.MySQL); err != nil {
		log.Fatalf("Failed to initialize MySQL: %v", err)
	}

	if err := database.AutoMigrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	fmt.Printf("Creating %d test users...\n", testUsers)

	users := make([]models.User, 0, testUsers)
	for i := 1; i <= testUsers; i++ {
		password, _ := auth.HashPassword(fmt.Sprintf("password%d", i))

		nicknames := []string{"张三", "李四", "王五", "Alice", "Bob", "ユーザー"}
		nickname := fmt.Sprintf("%s_%d", nicknames[i%len(nicknames)], i)

		users = append(users, models.User{
			UID:          fmt.Sprintf("uid_%d", i),
			Username:     fmt.Sprintf("user%d", i),
			PasswordHash: password,
			Nickname:     nickname,
			Avatar:       "",
			IsAvailable:  1,
		})
	}

	// 批量插入
	if err := database.DB.CreateInBatches(users, 50).Error; err != nil {
		log.Fatalf("Failed to insert users: %v", err)
	}

	fmt.Printf("✅ Successfully created %d test users\n", testUsers)
	fmt.Println("\n📝 Test accounts:")
	fmt.Println("  Username: user1 ~ user200")
	fmt.Println("  Password: password1 ~ password200")
	fmt.Println("  Example: user1 / password1")
	fmt.Println("\n🌐 Access:")
	fmt.Println("  Login page: http://localhost:8080/web/login.html")
}
