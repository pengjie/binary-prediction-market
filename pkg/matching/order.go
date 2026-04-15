package matching

import (
	"time"

	"github.com/huinong/golang-claude/pkg/common/decimal"
	"gorm.io/gorm"
)

// Direction indicates the direction of the order (buy/sell for YES or NO)
type Direction string

const (
	DirectionBuyYes  = Direction("BUY_YES")
	DirectionSellYes = Direction("SELL_YES")
	DirectionBuyNo   = Direction("BUY_NO")
	DirectionSellNo  = Direction("SELL_NO")
)

// Status represents the current status of the order
type Status string

const (
	StatusPending  = Status("PENDING")
	StatusFilled   = Status("FILLED")
	StatusCanceled = Status("CANCELED")
)

// Order represents a user's trading order in the prediction market.
// It tracks all details of a completed trade since we match immediately with the market maker.
type Order struct {
	ID         uint                `gorm:"primaryKey" json:"id"`
	UserID     uint                `json:"user_id"`
	EventID    uint                `json:"event_id"`
	Direction  Direction           `gorm:"size:20" json:"direction"`
	Quantity   decimal.Decimal     `gorm:"type:decimal(20,8)" json:"quantity"`
	Price      decimal.Decimal     `gorm:"type:decimal(10,8)" json:"price"`
	Amount     decimal.Decimal     `gorm:"type:decimal(20,8)" json:"amount"` // Total amount charged/credited
	Status     Status              `gorm:"size:20" json:"status"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
	DeletedAt  gorm.DeletedAt      `gorm:"index" json:"-"`
}
