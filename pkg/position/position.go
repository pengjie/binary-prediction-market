package position

import (
	"time"

	"github.com/huinong/golang-claude/pkg/common/decimal"
	"gorm.io/gorm"
)

// Position represents a user's position (holdings) in a specific event.
// It tracks how many YES and NO shares the user holds.
type Position struct {
	ID         uint                `gorm:"primaryKey" json:"id"`
	UserID     uint                `json:"user_id" gorm:"uniqueIndex:idx_user_event"`
	EventID    uint                `json:"event_id" gorm:"uniqueIndex:idx_user_event"`
	YesQuantity decimal.Decimal    `gorm:"type:decimal(20,8)" json:"yes_quantity"`
	NoQuantity  decimal.Decimal    `gorm:"type:decimal(20,8)" json:"no_quantity"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
	DeletedAt  gorm.DeletedAt      `gorm:"index" json:"-"`
}

// TotalCost calculates the total cost basis for this position.
// YES shares cost YesQuantity * entry price, but we don't track entry price here.
// For settlement, we only care about the final quantities.
func (p *Position) TotalQuantity() decimal.Decimal {
	return p.YesQuantity.Add(p.NoQuantity)
}
