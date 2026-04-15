// Package fund provides fund account management for the binary prediction market.
// It handles deposit, withdrawal, locking/unlocking funds for trading, and profit
// calculation after event settlement.
package fund

import (
	"time"

	"github.com/huinong/golang-claude/pkg/common/decimal"
	"gorm.io/gorm"
)

// Account represents a user's fund account that tracks balance and cumulative
// statistics for all trading activities. Each user has exactly one account.
type Account struct {
	ID            uint                `gorm:"primaryKey" json:"id"`
	UserID        uint                `gorm:"uniqueIndex" json:"user_id"`
	Balance       decimal.Decimal     `gorm:"type:decimal(20,8)" json:"balance"`
	TotalDeposit  decimal.Decimal     `gorm:"type:decimal(20,8)" json:"total_deposit"`
	TotalWithdraw decimal.Decimal     `gorm:"type:decimal(20,8)" json:"total_withdraw"`
	TotalPnL      decimal.Decimal     `gorm:"type:decimal(20,8)" json:"total_pnl"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
	DeletedAt     gorm.DeletedAt      `gorm:"index" json:"-"`
}
