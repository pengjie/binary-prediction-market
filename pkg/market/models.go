package market

import (
	"github.com/huinong/golang-claude/pkg/common/decimal"
	"github.com/huinong/golang-claude/pkg/event"
)

// MarketMakerParams contains all configurable parameters for the market maker.
type MarketMakerParams struct {
	BaseSpread   decimal.Decimal `json:"base_spread"`
	BaseSkew     decimal.Decimal `json:"base_skew"`
	Alpha        decimal.Decimal `json:"alpha"`   // Inventory ratio coefficient
	Beta         decimal.Decimal `json:"beta"`    // Probability coefficient
	Gamma        decimal.Decimal `json:"gamma"`   // Liquidity coefficient
	VolThreshold decimal.Decimal `json:"vol_threshold"` // Emergency break threshold
}

// DefaultParams returns default parameters for MVP.
func DefaultParams() MarketMakerParams {
	return MarketMakerParams{
		BaseSpread:   decimal.NewFromInt(1).Div(decimal.NewFromInt(100)),  // 1%
		BaseSkew:     decimal.Zero,
		Alpha:        decimal.NewFromInt(5).Div(decimal.NewFromInt(10)),  // 0.5
		Beta:         decimal.NewFromInt(3).Div(decimal.NewFromInt(10)),  // 0.3
		Gamma:        decimal.NewFromInt(2).Div(decimal.NewFromInt(10)), // 0.2
		VolThreshold: decimal.NewFromInt(10).Div(decimal.NewFromInt(100)), // 10%
	}
}

// PriceQuotation contains the final bid/ask prices after all adjustments.
type PriceQuotation struct {
	YesBid  decimal.Decimal `json:"yes_bid"`  // Platform buys YES at this price
	YesAsk  decimal.Decimal `json:"yes_ask"`  // Platform sells YES at this price
	NoBid   decimal.Decimal `json:"no_bid"`   // Platform buys NO at this price
	NoAsk   decimal.Decimal `json:"no_ask"`   // Platform sells NO at this price
}

// MarketState contains current state of the market for an event.
type MarketState struct {
	Event         *event.Event    `json:"event"`
	ShadowPrice   decimal.Decimal `json:"shadow_price"`
	FinalYesPrice decimal.Decimal `json:"final_yes_price"`
	Skew          decimal.Decimal `json:"skew"`
	Spread        decimal.Decimal `json:"spread"`
}
