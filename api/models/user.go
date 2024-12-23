package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user account
type User struct {
	gorm.Model
	ID        string    `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex"`
	Password  string    `json:"password,omitempty"`
	Role      string    `json:"role"`
	LastLogin time.Time `json:"last_login"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
