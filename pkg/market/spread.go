package market

import (
	"time"

	"github.com/huinong/golang-claude/pkg/common/decimal"
	"github.com/huinong/golang-claude/pkg/event"
)

// CalculateSpread adjusts spread based on time to maturity and liquidity.
// Spread increases as event approaches end (breathing effect).
// Closer to expiration = higher spread = lower liquidity.
func CalculateSpread(
	baseSpread decimal.Decimal, e *event.Event) decimal.Decimal {
	now := time.Now()
	if now.After(e.EndTime) {
		return baseSpread.Mul(decimal.NewFromInt(2))
	}

	// Calculate time multiplier: closer to end = larger spread
	durationTotal := e.EndTime.Sub(e.StartTime)
	timeRemaining := e.EndTime.Sub(now)
	if durationTotal <= 0 {
		return baseSpread.Mul(decimal.NewFromInt(2))
	}

	// 0 ~ 3x scaling: when 1 month out = 0.3%, 1 day out = 1.0%
	timeRatio := decimal.NewFromFloat(float64(timeRemaining) / float64(durationTotal))
	// minimum multiplier is 1.0, maximum is 3.0
	multiplier := decimal.One.Add(decimal.Two.Mul(decimal.One.Sub(timeRatio)))

	return baseSpread.Mul(multiplier)
}
