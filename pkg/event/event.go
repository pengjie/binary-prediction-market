// Package event provides binary outcome event management for the prediction market.
// Events represent prediction questions with two possible outcomes (YES/NO)
// that can be traded and settled after the event end time.
package event

import (
	"time"

	"github.com/huinong/golang-claude/pkg/common/decimal"
	"gorm.io/gorm"
)

// EventStatus represents the current status of an event.
type EventStatus string

const (
	// StatusPending means the event is created but trading hasn't started.
	StatusPending = EventStatus("PENDING")
	// StatusTrading means the event is active and open for trading.
	StatusTrading = EventStatus("TRADING")
	// StatusSettled means the event has ended and result is determined.
	StatusSettled = EventStatus("SETTLED")
)

// Result represents the outcome of a settled binary event.
type Result string

const (
	// ResultYesWon means the YES outcome won.
	ResultYesWon = Result("YES_WON")
	// ResultNoWon means the NO outcome won.
	ResultNoWon = Result("NO_WON")
)

// Event represents a binary prediction market event with two possible outcomes.
type Event struct {
	ID            uint                `gorm:"primaryKey" json:"id"`
	Title         string              `gorm:"size:200" json:"title"`
	Description   string              `gorm:"size:1000" json:"description"`
	StartTime     time.Time           `json:"start_time"`
	EndTime       time.Time           `json:"end_time"`
	Status        EventStatus         `gorm:"size:20" json:"status"`
	Result        Result              `gorm:"size:20" json:"result,omitempty"`
	YesPrice      decimal.Decimal     `gorm:"type:decimal(10,8)" json:"yes_price"` // 0~1 probability
	YesInventory  decimal.Decimal     `gorm:"type:decimal(20,8)" json:"yes_inventory"`
	NoInventory   decimal.Decimal     `gorm:"type:decimal(20,8)" json:"no_inventory"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
	DeletedAt     gorm.DeletedAt      `gorm:"index" json:"-"`
}

// NoPrice returns 1 - YesPrice (price of NO outcome).
func (e *Event) NoPrice() decimal.Decimal {
	return decimal.One.Sub(e.YesPrice)
}

// TotalInventory returns the sum of YES and NO inventory (should equal initial supply).
func (e *Event) TotalInventory() decimal.Decimal {
	return e.YesInventory.Add(e.NoInventory)
}
