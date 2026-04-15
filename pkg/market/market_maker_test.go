package market

import (
	"testing"
	"time"

	"github.com/huinong/golang-claude/pkg/common/decimal"
	"github.com/huinong/golang-claude/pkg/event"
)

func TestCalculateQuotation(t *testing.T) {
	e := &event.Event{
		YesPrice:     decimal.NewFromInt(60).Div(decimal.NewFromInt(100)),
		YesInventory: decimal.NewFromInt(400),
		NoInventory:  decimal.NewFromInt(600),
		StartTime:    time.Now().Add(-24 * time.Hour),
		EndTime:      time.Now().Add(24 * time.Hour),
	}

	params := DefaultParams()
	pe := NewSimplePriceEngine()
	mm := NewMarketMaker(params, pe)

	quote := mm.CalculateQuotation(e, decimal.NewFromInt(1))

	// Spread should be positive with ask always higher than bid
	if !quote.YesAsk.GreaterThan(quote.YesBid) {
		t.Errorf("yes_ask %v should be greater than yes_bid %v", quote.YesAsk, quote.YesBid)
	}
	if !quote.NoAsk.GreaterThan(quote.NoBid) {
		t.Errorf("no_ask %v should be greater than no_bid %v", quote.NoAsk, quote.NoBid)
	}

	// Sum of complementary prices should be approximately 1
	sumBid := quote.YesBid.Add(quote.NoAsk)
	if !sumBid.Round(4).Equal(decimal.One.Round(4)) {
		t.Errorf("yes_bid + no_ask should be ~1, got %v", sumBid)
	}
	sumAsk := quote.YesAsk.Add(quote.NoBid)
	if !sumAsk.Round(4).Equal(decimal.One.Round(4)) {
		t.Errorf("yes_ask + no_bid should be ~1, got %v", sumAsk)
	}

	// All prices should be between 0.01 and 0.99
	prices := []decimal.Decimal{quote.YesBid, quote.YesAsk, quote.NoBid, quote.NoAsk}
	minPrice := decimal.NewFromInt(1).Div(decimal.NewFromInt(100))
	maxPrice := decimal.NewFromInt(99).Div(decimal.NewFromInt(100))
	for _, p := range prices {
		if p.LessThan(minPrice) || p.GreaterThan(maxPrice) {
			t.Errorf("price %v out of valid range [0.01, 0.99]", p)
		}
	}
}

func TestSkewWhenInventoryImbalanced(t *testing.T) {
	// 900 YES inventory, 100 NO → heavy imbalance, should have positive skew
	// Positive skew pushes YES price up, which encourages users to sell YES to the platform
	// This helps reduce the imbalance (platform has too much YES, needs users to take it off their hands)
	e := &event.Event{
		YesPrice:     decimal.NewFromInt(50).Div(decimal.NewFromInt(100)),
		YesInventory: decimal.NewFromInt(900),
		NoInventory:  decimal.NewFromInt(100),
	}

	params := DefaultParams()
	skew := CalculateSkew(params, e, decimal.NewFromInt(1))

	// Should be positive because inventory is YES-heavy
	if skew.LessThan(decimal.Zero) {
		t.Errorf("expected positive skew for YES-heavy inventory, got %v", skew)
	}
}

func TestSkewWhenReverseImbalanced(t *testing.T) {
	// 100 YES inventory, 900 NO → reverse imbalance, should have negative skew
	// Negative skew pushes YES price down, encourages users to buy YES to balance inventory
	e := &event.Event{
		YesPrice:     decimal.NewFromInt(50).Div(decimal.NewFromInt(100)),
		YesInventory: decimal.NewFromInt(100),
		NoInventory:  decimal.NewFromInt(900),
	}

	params := DefaultParams()
	skew := CalculateSkew(params, e, decimal.NewFromInt(1))

	// Should be negative because inventory is NO-heavy
	if skew.GreaterThan(decimal.Zero) {
		t.Errorf("expected negative skew for NO-heavy inventory, got %v", skew)
	}
}

func TestSkewWhenBalanced(t *testing.T) {
	// 500 YES, 500 NO → perfectly balanced, skew should be close to base skew
	e := &event.Event{
		YesPrice:     decimal.NewFromInt(50).Div(decimal.NewFromInt(100)),
		YesInventory: decimal.NewFromInt(500),
		NoInventory:  decimal.NewFromInt(500),
	}

	params := DefaultParams()
	skew := CalculateSkew(params, e, decimal.NewFromInt(1))

	// With balanced inventory and 50% price, skew should be:
	// skew = 0 + 0.5*0 + 0.3*(1-0.5) + 0.2*1 = 0.15 + 0.2 = 0.35
	expected := params.Beta.Mul(decimal.NewFromInt(50).Div(decimal.NewFromInt(100))).Add(params.Gamma)
	if !skew.Round(3).Equal(expected.Round(3)) {
		t.Errorf("expected skew ~%v for balanced inventory, got %v", expected, skew)
	}
}

func TestCalculateSpreadIncreasesNearEnd(t *testing.T) {
	// Event with one day duration, 1 hour remaining
	start := time.Now().Add(-23 * time.Hour)
	end := time.Now().Add(1 * time.Hour)
	e := &event.Event{
		StartTime: start,
		EndTime:   end,
	}

	baseSpread := decimal.NewFromInt(1).Div(decimal.NewFromInt(100)) // 1%
	spread := CalculateSpread(baseSpread, e)

	// Spread should be larger than base because we're close to end
	if !spread.GreaterThan(baseSpread) {
		t.Errorf("expected spread larger than base %v near expiration, got %v", baseSpread, spread)
	}

	// Maximum spread should be 3x base
	maxSpread := baseSpread.Mul(decimal.NewFromInt(3))
	if spread.GreaterThan(maxSpread) {
		t.Errorf("expected spread no more than 3x base %v, got %v", maxSpread, spread)
	}
}

func TestAnchorPrice(t *testing.T) {
	current := decimal.NewFromInt(55).Div(decimal.NewFromInt(100)) // 0.55
	shadow := decimal.NewFromInt(50).Div(decimal.NewFromInt(100))  // 0.50
	strength := decimal.NewFromInt(2).Div(decimal.NewFromInt(10))    // 0.2 strength

	result := AnchorPrice(current, shadow, strength)
	// 0.55 * 0.8 + 0.50 * 0.2 = 0.44 + 0.10 = 0.54
	expected := decimal.NewFromInt(54).Div(decimal.NewFromInt(100))

	if !result.Equal(expected) {
		t.Errorf("expected anchored price %v, got %v", expected, result)
	}
}

func TestGetState(t *testing.T) {
	e := &event.Event{
		YesPrice:     decimal.NewFromInt(60).Div(decimal.NewFromInt(100)),
		YesInventory: decimal.NewFromInt(400),
		NoInventory:  decimal.NewFromInt(600),
		StartTime:    time.Now().Add(-24 * time.Hour),
		EndTime:      time.Now().Add(24 * time.Hour),
	}

	params := DefaultParams()
	pe := NewSimplePriceEngine()
	mm := NewMarketMaker(params, pe)

	state := mm.GetState(e, decimal.NewFromInt(1))

	if state.Event != e {
		t.Errorf("expected event pointer to be preserved")
	}
	if state.ShadowPrice != e.YesPrice {
		t.Errorf("expected shadow price to equal event yes price")
	}
	if state.FinalYesPrice.LessThan(decimal.NewFromInt(1).Div(decimal.NewFromInt(100))) ||
		state.FinalYesPrice.GreaterThan(decimal.NewFromInt(99).Div(decimal.NewFromInt(100))) {
		t.Errorf("final price %v out of valid range", state.FinalYesPrice)
	}
}
