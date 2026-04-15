package market

import (
	"github.com/huinong/golang-claude/pkg/common/decimal"
)

// AnchorPrice anchors the adjusted price toward shadow price.
// Keeps market price from drifting too far from the theoretical probability.
// Strength determines how much weight to give the shadow price:
// - strength 0.2 means 20% shadow price, 80% current price
// - strength closer to 1 means more anchoring force
func AnchorPrice(
	currentPrice decimal.Decimal, shadowPrice decimal.Decimal, strength decimal.Decimal) decimal.Decimal {
	// weighted average: currentPrice * (1 - strength) + shadowPrice * strength
	return currentPrice.Mul(decimal.One.Sub(strength)).Add(shadowPrice.Mul(strength))
}
