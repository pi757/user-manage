package models

import (
	"time"
)

// User 用户模型
type User struct {
	ID           uint      `gorm:"column:id" json:"id"`
	UID          string    `gorm:"column:uid" json:"uid"`
	Username     string    `gorm:"column:username" json:"username"`
	Nickname     string    `gorm:"column:nickname" json:"nickname"`
	PasswordHash string    `gorm:"column:password_hash" json:"-"`
	Avatar       string    `gorm:"column:avatar" json:"avatar"`
	IsAvailable  int8      `gorm:"column:is_available" json:"is_available"`
	CreateTime   time.Time `gorm:"column:create_time;autoCreateTime" json:"create_time"`
	UpdateTime   time.Time `gorm:"column:update_time;autoUpdateTime" json:"update_time"`
}
