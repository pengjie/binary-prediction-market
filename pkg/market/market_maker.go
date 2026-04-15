package market

import (
	"github.com/huinong/golang-claude/pkg/common/decimal"
	"github.com/huinong/golang-claude/pkg/event"
)

// MarketMaker is the core of the automated market maker.
// It implements the three-force model:
//   - Spread (breathing): adjusts spread based on time to maturity
//   - Skew (gravitational): adjusts price based on inventory imbalance
//   - Anchor (centripetal): anchors price to theoretical probability
type MarketMaker struct {
	params      MarketMakerParams
	priceEngine PriceEngine
}

// NewMarketMaker creates a new MarketMaker with the given parameters and price engine.
func NewMarketMaker(params MarketMakerParams, priceEngine PriceEngine) *MarketMaker {
	return &MarketMaker{
		params:      params,
		priceEngine: priceEngine,
	}
}

// CalculateQuotation calculates the final bid/ask prices after all adjustments.
// Returns the four prices that the market maker is willing to trade at:
// - YesBid: platform buys YES at this price
// - YesAsk: platform sells YES at this price
// - NoBid: platform buys NO at this price
// - NoAsk: platform sells NO at this price
func (mm *MarketMaker) CalculateQuotation(e *event.Event, liquidity decimal.Decimal) *PriceQuotation {
	shadowPrice := mm.priceEngine.CalculateShadowPrice(e)
	spread := CalculateSpread(mm.params.BaseSpread, e)
	skew := CalculateSkew(mm.params, e, liquidity)

	// Apply skew to shadow price
	adjustedPrice := shadowPrice.Add(skew)

	// Clamp to 0.01 - 0.99 to avoid 0 or 1 which causes invalid prices
	minPrice := decimal.NewFromInt(1).Div(decimal.NewFromInt(100))
	maxPrice := decimal.NewFromInt(99).Div(decimal.NewFromInt(100))
	if adjustedPrice.LessThan(minPrice) {
		adjustedPrice = minPrice
	}
	if adjustedPrice.GreaterThan(maxPrice) {
		adjustedPrice = maxPrice
	}

	halfSpread := spread.Div(decimal.NewFromInt(2))

	// Yes: bid/ask
	yesBid := adjustedPrice.Sub(halfSpread)
	yesAsk := adjustedPrice.Add(halfSpread)

	// No prices are symmetric: yes + no = 1, so same spread is inverted
	noAsk := decimal.One.Sub(yesBid)
	noBid := decimal.One.Sub(yesAsk)

	return &PriceQuotation{
		YesBid: yesBid,
		YesAsk: yesAsk,
		NoBid:  noBid,
		NoAsk:  noAsk,
	}
}

// CheckEmergencyBreak checks if we need to widen spread due to high volatility.
func (mm *MarketMaker) CheckEmergencyBreak(volatility decimal.Decimal) bool {
	return volatility.GreaterThan(mm.params.VolThreshold)
}

// GetState returns current market state for debugging/inspection.
func (mm *MarketMaker) GetState(e *event.Event, liquidity decimal.Decimal) *MarketState {
	shadowPrice := mm.priceEngine.CalculateShadowPrice(e)
	spread := CalculateSpread(mm.params.BaseSpread, e)
	skew := CalculateSkew(mm.params, e, liquidity)
	finalPrice := shadowPrice.Add(skew)

	// Clamp same as in CalculateQuotation
	minPrice := decimal.NewFromInt(1).Div(decimal.NewFromInt(100))
	maxPrice := decimal.NewFromInt(99).Div(decimal.NewFromInt(100))
	if finalPrice.LessThan(minPrice) {
		finalPrice = minPrice
	}
	if finalPrice.GreaterThan(maxPrice) {
		finalPrice = maxPrice
	}

	return &MarketState{
		Event:         e,
		ShadowPrice:   shadowPrice,
		FinalYesPrice: finalPrice,
		Skew:          skew,
		Spread:        spread,
	}
}
