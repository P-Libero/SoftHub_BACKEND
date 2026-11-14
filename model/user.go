package model

import "time"

type User struct {
	ID           uint64    `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"size:50;not null;uniqueIndex:idx_user_username"`
	PasswordHash string    `json:"-" gorm:"column:password_hash;size:255;not null"`
	Role         string    `json:"role" gorm:"size:20;not null;default:user"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "user"
}
