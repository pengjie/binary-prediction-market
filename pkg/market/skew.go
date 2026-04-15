package market

import (
	"github.com/huinong/golang-claude/pkg/common/decimal"
	"github.com/huinong/golang-claude/pkg/event"
)

// CalculateSkew calculates price skew based on inventory imbalance.
// Skew adjusts price to incentivize trading that reduces inventory imbalance.
// Formula: Skew = base_skew + α * Inventory_Ratio + β * (1 - Event_Probability) + γ * Liquidity
// Where:
//   - Inventory_Ratio = (YesInventory - NoInventory) / TotalInventory
//   - α = sensitivity to inventory imbalance
//   - β = sensitivity to probability deviation from 50%
//   - γ = sensitivity to recent liquidity/volume
func CalculateSkew(
	params MarketMakerParams,
	e *event.Event,
	liquidity decimal.Decimal,
) decimal.Decimal {
	totalInv := e.TotalInventory()
	if totalInv.IsZero() {
		return params.BaseSkew
	}

	invRatio := e.YesInventory.Sub(e.NoInventory).Div(totalInv)

	term1 := params.Alpha.Mul(invRatio)
	term2 := params.Beta.Mul(decimal.One.Sub(e.YesPrice))
	term3 := params.Gamma.Mul(liquidity)

	return params.BaseSkew.Add(term1).Add(term2).Add(term3)
}
