package matching

import (
	"github.com/huinong/golang-claude/pkg/common/decimal"
	"github.com/huinong/golang-claude/pkg/common/errors"
	"github.com/huinong/golang-claude/pkg/event"
	"github.com/huinong/golang-claude/pkg/fund"
	"github.com/huinong/golang-claude/pkg/market"
	"github.com/huinong/golang-claude/pkg/position"
)

// MatchingEngine handles order matching and execution with the automated market maker.
// It coordinates between funds, positions, event inventory and the market maker.
type MatchingEngine struct {
	eventService    event.Service
	fundService     fund.Service
	positionService position.Service
	marketMaker     *market.MarketMaker
}

// NewMatchingEngine creates a new MatchingEngine with the required dependencies.
func NewMatchingEngine(
	es event.Service,
	fs fund.Service,
	ps position.Service,
	mm *market.MarketMaker,
) *MatchingEngine {
	return &MatchingEngine{
		eventService:    es,
		fundService:     fs,
		positionService: ps,
		marketMaker:     mm,
	}
}

// BuyYes buys YES shares from the platform.
// User pays quantity * price to platform, receives YES shares.
// Platform YES inventory decreases, NO inventory increases.
func (me *MatchingEngine) BuyYes(userID uint, eventID uint, quantity decimal.Decimal) (*Order, error) {
	e, err := me.eventService.GetByID(eventID)
	if err != nil {
		return nil, err
	}
	if e.Status != event.StatusTrading {
		return nil, errors.New(errors.CodeBadRequest, "event not trading")
	}

	quote := me.marketMaker.CalculateQuotation(e, decimal.NewFromInt(1))
	price := quote.YesAsk
	totalAmount := quantity.Mul(price)

	// Lock funds from user
	err = me.fundService.Lock(userID, totalAmount)
	if err != nil {
		return nil, err
	}

	// Update platform inventory: we sold YES to user, so YES inventory decreases
	// We received payment: increase NO inventory (platform takes opposite position)
	e.YesInventory = e.YesInventory.Sub(quantity)
	e.NoInventory = e.NoInventory.Add(quantity)
	err = me.eventService.UpdatePrice(eventID, e.YesPrice)
	if err != nil {
		// Rollback lock
		me.fundService.Unlock(userID, totalAmount)
		return nil, err
	}

	// Add position to user
	err = me.positionService.AddPosition(userID, eventID, quantity, position.DirectionYes)
	if err != nil {
		// Rollback: unlock funds back to user
		me.fundService.Unlock(userID, totalAmount)
		e.YesInventory = e.YesInventory.Add(quantity)
		e.NoInventory = e.NoInventory.Sub(quantity)
		me.eventService.UpdatePrice(eventID, e.YesPrice)
		return nil, err
	}

	// Trade is complete: funds already locked by platform kept as payment

	// Create order record
	order := &Order{
		UserID:    userID,
		EventID:   eventID,
		Direction: DirectionBuyYes,
		Quantity:  quantity,
		Price:     price,
		Amount:    totalAmount,
		Status:    StatusFilled,
	}

	return order, nil
}

// SellYes sells YES shares to the platform.
// User receives quantity * price from platform, sells YES shares back to platform.
// Platform YES inventory increases, NO inventory decreases.
func (me *MatchingEngine) SellYes(userID uint, eventID uint, quantity decimal.Decimal) (*Order, error) {
	e, err := me.eventService.GetByID(eventID)
	if err != nil {
		return nil, err
	}
	if e.Status != event.StatusTrading {
		return nil, errors.New(errors.CodeBadRequest, "event not trading")
	}

	// Check user has enough YES shares
	pos, err := me.positionService.GetPosition(userID, eventID)
	if err != nil {
		return nil, err
	}
	if pos.YesQuantity.Cmp(quantity) < 0 {
		return nil, errors.New(errors.CodeInsufficientInventory, "insufficient YES shares")
	}

	quote := me.marketMaker.CalculateQuotation(e, decimal.NewFromInt(1))
	price := quote.YesBid
	totalAmount := quantity.Mul(price)

	// Lock the shares (already done by checking position, but we need to unlock on failure)
	// Credit proceeds to user after execution

	// Update platform inventory: we bought YES from user, so YES inventory increases
	e.YesInventory = e.YesInventory.Add(quantity)
	e.NoInventory = e.NoInventory.Sub(quantity)
	err = me.eventService.UpdatePrice(eventID, e.YesPrice)
	if err != nil {
		return nil, err
	}

	// Remove shares from user position (add to platform, user sells)
	// In our position model, we just add - but actually we just track net position
	err = me.positionService.AddPosition(userID, eventID, quantity, position.DirectionNo)
	if err != nil {
		// Rollback inventory
		e.YesInventory = e.YesInventory.Sub(quantity)
		e.NoInventory = e.NoInventory.Add(quantity)
		me.eventService.UpdatePrice(eventID, e.YesPrice)
		return nil, err
	}

	// Add proceeds to user balance
	err = me.fundService.AddProfit(userID, totalAmount)
	if err != nil {
		// Rollback everything
		e.YesInventory = e.YesInventory.Sub(quantity)
		e.NoInventory = e.NoInventory.Add(quantity)
		me.eventService.UpdatePrice(eventID, e.YesPrice)
		return nil, err
	}

	// Create order record
	order := &Order{
		UserID:    userID,
		EventID:   eventID,
		Direction: DirectionSellYes,
		Quantity:  quantity,
		Price:     price,
		Amount:    totalAmount,
		Status:    StatusFilled,
	}

	return order, nil
}

// BuyNo buys NO shares from the platform.
// User pays quantity * price to platform, receives NO shares.
// Platform NO inventory decreases, YES inventory increases.
func (me *MatchingEngine) BuyNo(userID uint, eventID uint, quantity decimal.Decimal) (*Order, error) {
	e, err := me.eventService.GetByID(eventID)
	if err != nil {
		return nil, err
	}
	if e.Status != event.StatusTrading {
		return nil, errors.New(errors.CodeBadRequest, "event not trading")
	}

	quote := me.marketMaker.CalculateQuotation(e, decimal.NewFromInt(1))
	price := quote.NoAsk
	totalAmount := quantity.Mul(price)

	// Lock funds from user
	err = me.fundService.Lock(userID, totalAmount)
	if err != nil {
		return nil, err
	}

	// Update platform inventory: we sold NO to user, so NO inventory decreases
	// We received payment: increase YES inventory (platform takes opposite position)
	e.NoInventory = e.NoInventory.Sub(quantity)
	e.YesInventory = e.YesInventory.Add(quantity)
	err = me.eventService.UpdatePrice(eventID, e.YesPrice)
	if err != nil {
		me.fundService.Unlock(userID, totalAmount)
		return nil, err
	}

	// Add position to user
	err = me.positionService.AddPosition(userID, eventID, quantity, position.DirectionNo)
	if err != nil {
		// Rollback: unlock funds back to user
		me.fundService.Unlock(userID, totalAmount)
		e.NoInventory = e.NoInventory.Add(quantity)
		e.YesInventory = e.YesInventory.Sub(quantity)
		me.eventService.UpdatePrice(eventID, e.YesPrice)
		return nil, err
	}

	// Trade is complete: funds already locked by platform kept as payment

	// Create order record
	order := &Order{
		UserID:    userID,
		EventID:   eventID,
		Direction: DirectionBuyNo,
		Quantity:  quantity,
		Price:     price,
		Amount:    totalAmount,
		Status:    StatusFilled,
	}

	return order, nil
}

// SellNo sells NO shares to the platform.
// User receives quantity * price from platform, sells NO shares back to platform.
// Platform NO inventory increases, YES inventory decreases.
func (me *MatchingEngine) SellNo(userID uint, eventID uint, quantity decimal.Decimal) (*Order, error) {
	e, err := me.eventService.GetByID(eventID)
	if err != nil {
		return nil, err
	}
	if e.Status != event.StatusTrading {
		return nil, errors.New(errors.CodeBadRequest, "event not trading")
	}

	// Check user has enough NO shares
	pos, err := me.positionService.GetPosition(userID, eventID)
	if err != nil {
		return nil, err
	}
	if pos.NoQuantity.Cmp(quantity) < 0 {
		return nil, errors.New(errors.CodeInsufficientInventory, "insufficient NO shares")
	}

	quote := me.marketMaker.CalculateQuotation(e, decimal.NewFromInt(1))
	price := quote.NoBid
	totalAmount := quantity.Mul(price)

	// Update platform inventory: we bought NO from user, so NO inventory increases
	e.NoInventory = e.NoInventory.Add(quantity)
	e.YesInventory = e.YesInventory.Sub(quantity)
	err = me.eventService.UpdatePrice(eventID, e.YesPrice)
	if err != nil {
		return nil, err
	}

	// Remove shares from user position
	err = me.positionService.AddPosition(userID, eventID, quantity, position.DirectionYes)
	if err != nil {
		// Rollback inventory
		e.NoInventory = e.NoInventory.Sub(quantity)
		e.YesInventory = e.YesInventory.Add(quantity)
		me.eventService.UpdatePrice(eventID, e.YesPrice)
		return nil, err
	}

	// Add proceeds to user balance
	err = me.fundService.AddProfit(userID, totalAmount)
	if err != nil {
		// Rollback everything
		e.NoInventory = e.NoInventory.Sub(quantity)
		e.YesInventory = e.YesInventory.Add(quantity)
		me.eventService.UpdatePrice(eventID, e.YesPrice)
		return nil, err
	}

	// Create order record
	order := &Order{
		UserID:    userID,
		EventID:   eventID,
		Direction: DirectionSellNo,
		Quantity:  quantity,
		Price:     price,
		Amount:    totalAmount,
		Status:    StatusFilled,
	}

	return order, nil
}

// GetOrder is a placeholder for future implementation if we need to retrieve orders by ID.
// For MVP, we don't need full order listing capability.
