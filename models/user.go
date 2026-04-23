package models

import (
	"time"
)

// User 用户模型
type User struct {
	ID           uint      `gorm:"primarykey;column:id" json:"id"`
	UID          string    `gorm:"column:uid;size:64;uniqueIndex;not null;default:''" json:"uid"`
	Username     string    `gorm:"column:username;size:64;uniqueIndex;not null;default:''" json:"username"`
	Nickname     string    `gorm:"column:nickname;size:64;not null;default:''" json:"nickname"`
	PasswordHash string    `gorm:"column:password_hash;size:128;not null;default:''" json:"-"`
	Avatar       string    `gorm:"column:avatar;size:256;not null;default:''" json:"avatar"`
	IsAvailable  int8      `gorm:"column:is_available;type:tinyint;default:1" json:"is_available"`
	CreateTime   time.Time `gorm:"column:create_time;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"create_time"`
	UpdateTime   time.Time `gorm:"column:update_time;type:timestamp;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"update_time"`
}
