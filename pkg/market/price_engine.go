package market

import (
	"github.com/huinong/golang-claude/pkg/common/decimal"
	"github.com/huinong/golang-claude/pkg/event"
)

// PriceEngine calculates the shadow price based on current conditions.
type PriceEngine interface {
	CalculateShadowPrice(event *event.Event) decimal.Decimal
}

// SimplePriceEngine is a simple implementation that uses the current event price.
// In future versions, this can be updated based on order flow to reflect
// changing market probabilities.
type SimplePriceEngine struct {
}

// NewSimplePriceEngine creates a new SimplePriceEngine.
func NewSimplePriceEngine() PriceEngine {
	return &SimplePriceEngine{}
}

// CalculateShadowPrice returns the current event's yes price.
func (e *SimplePriceEngine) CalculateShadowPrice(event *event.Event) decimal.Decimal {
	return event.YesPrice
}
