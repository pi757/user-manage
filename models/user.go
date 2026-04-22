package models

import (
	"gorm.io/gorm"
	"time"
)

// User 用户模型
type User struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	Username     string         `gorm:"size:64;uniqueIndex;not null" json:"username"`
	Password     string         `gorm:"size:255;not null" json:"-"`
	Nickname     string         `gorm:"size:128" json:"nickname"`
	Avatar       string         `gorm:"size:255" json:"avatar"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// Session Session模型
type Session struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Token     string    `gorm:"size:64;uniqueIndex;not null" json:"token"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (Session) TableName() string {
	return "sessions"
}
