package matching

import (
	"github.com/huinong/golang-claude/pkg/common/decimal"
	"github.com/huinong/golang-claude/pkg/event"
	"github.com/huinong/golang-claude/pkg/fund"
	"github.com/huinong/golang-claude/pkg/position"
)

// SettleEvent settles an event after the result is known and distributes payouts to all users.
// Each YES share pays out 1 unit if YES won, each NO share pays 1 unit if NO won.
func SettleEvent(
	e *event.Event,
	ps position.Service,
	fs fund.Service,
) error {
	// Get all positions for this event
	positions, err := ps.ListByEvent(e.ID)
	if err != nil {
		return err
	}

	// Distribute payout to each position
	for _, pos := range positions {
		var payout decimal.Decimal
		if e.Result == event.ResultYesWon {
			// Each YES share pays out 1 unit
			payout = pos.YesQuantity.Mul(decimal.One)
		} else {
			// Each NO share pays out 1 unit
			payout = pos.NoQuantity.Mul(decimal.One)
		}

		// Add payout to user's account
		if payout.GreaterThan(decimal.Zero) {
			if err := fs.AddProfit(pos.UserID, payout); err != nil {
				return err
			}
		}
	}

	return nil
}
