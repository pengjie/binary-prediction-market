package user

import (
	"time"

	"gorm.io/gorm"
)

// User represents a platform user with API key authentication.
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Username  string         `gorm:"uniqueIndex;size:50" json:"username"`
	APIKey    string         `gorm:"uniqueIndex;size:100" json:"api_key"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
