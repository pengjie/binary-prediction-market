package matching

import (
	"testing"
	"time"

	"github.com/huinong/golang-claude/pkg/common/decimal"
	"github.com/huinong/golang-claude/pkg/event"
	"github.com/huinong/golang-claude/pkg/fund"
	"github.com/huinong/golang-claude/pkg/market"
	"github.com/huinong/golang-claude/pkg/position"
	"github.com/huinong/golang-claude/pkg/user"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	// Auto migrate all schemas
	err = db.AutoMigrate(
		&user.User{},
		&fund.Account{},
		&event.Event{},
		&position.Position{},
		&Order{},
	)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestBuyYes(t *testing.T) {
	db := setupTestDB(t)

	// Setup dependencies
	userRepo := user.NewGormRepository(db)
	userService := user.NewUserService(userRepo)
	u, err := userService.Create("testuser")
	if err != nil {
		t.Fatal(err)
	}

	fundRepo := fund.NewGormRepository(db)
	fundService := fund.NewService(fundRepo)
	err = fundService.CreateAccount(u)
	if err != nil {
		t.Fatal(err)
	}
	err = fundService.Deposit(u.ID, decimal.NewFromInt(1000))
	if err != nil {
		t.Fatal(err)
	}

	eventRepo := event.NewGormRepository(db)
	eventService := event.NewService(eventRepo)
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now().Add(1 * time.Hour)
	e, err := eventService.Create(
		"Test Event",
		"Test Description",
		startTime, endTime,
		decimal.NewFromInt(60).Div(decimal.NewFromInt(100)), // 60%
		decimal.NewFromInt(1000), // 1000 shares
	)
	if err != nil {
		t.Fatal(err)
	}
	err = eventService.StartTrading(e.ID)
	if err != nil {
		t.Fatal(err)
	}

	posRepo := position.NewGormRepository(db)
	posService := position.NewService(posRepo)

	params := market.DefaultParams()
	pe := market.NewSimplePriceEngine()
	mm := market.NewMarketMaker(params, pe)

	me := NewMatchingEngine(eventService, fundService, posService, mm)

	// Buy 10 YES shares
	order, err := me.BuyYes(u.ID, e.ID, decimal.NewFromInt(10))
	if err != nil {
		t.Fatalf("BuyYes failed: %v", err)
	}

	if order.Status != StatusFilled {
		t.Errorf("expected status FILLED, got %v", order.Status)
	}
	if !order.Quantity.Equal(decimal.NewFromInt(10)) {
		t.Errorf("expected quantity 10, got %v", order.Quantity)
	}

	// Check user balance after purchase
	acc, err := fundService.GetAccount(u.ID)
	if err != nil {
		t.Fatal(err)
	}
	// Price is about 0.6 + spread ~0.01, so total around 6.05
	expectedBalance := decimal.NewFromInt(1000).Sub(order.Amount)
	if !acc.Balance.Round(2).Equal(expectedBalance.Round(2)) {
		t.Errorf("expected balance ~%v, got %v", expectedBalance, acc.Balance)
	}

	// Check event inventory changed
	eUpdated, err := eventService.GetByID(e.ID)
	if err != nil {
		t.Fatal(err)
	}
	// Original: YesInventory 400, NoInventory 600
	// After buying 10 YES: YesInventory 400 - 10 = 390, NoInventory 600 + 10 = 610
	expectedYesInv := decimal.NewFromInt(400).Sub(decimal.NewFromInt(10))
	if !eUpdated.YesInventory.Equal(expectedYesInv) {
		t.Errorf("expected yes inventory %v, got %v", expectedYesInv, eUpdated.YesInventory)
	}
	expectedNoInv := decimal.NewFromInt(600).Add(decimal.NewFromInt(10))
	if !eUpdated.NoInventory.Equal(expectedNoInv) {
		t.Errorf("expected no inventory %v, got %v", expectedNoInv, eUpdated.NoInventory)
	}

	// Check user position
	userPos, err := posService.GetPosition(u.ID, e.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !userPos.YesQuantity.Equal(decimal.NewFromInt(10)) {
		t.Errorf("expected yes quantity 10, got %v", userPos.YesQuantity)
	}
}

func TestBuyNo(t *testing.T) {
	db := setupTestDB(t)

	// Setup dependencies
	userRepo := user.NewGormRepository(db)
	userService := user.NewUserService(userRepo)
	u, err := userService.Create("testuser2")
	if err != nil {
		t.Fatal(err)
	}

	fundRepo := fund.NewGormRepository(db)
	fundService := fund.NewService(fundRepo)
	err = fundService.CreateAccount(u)
	if err != nil {
		t.Fatal(err)
	}
	err = fundService.Deposit(u.ID, decimal.NewFromInt(1000))
	if err != nil {
		t.Fatal(err)
	}

	eventRepo := event.NewGormRepository(db)
	eventService := event.NewService(eventRepo)
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now().Add(1 * time.Hour)
	e, err := eventService.Create(
		"Test Event 2",
		"Test Description",
		startTime, endTime,
		decimal.NewFromInt(60).Div(decimal.NewFromInt(100)), // 60% YES
		decimal.NewFromInt(1000), // 1000 shares
	)
	if err != nil {
		t.Fatal(err)
	}
	err = eventService.StartTrading(e.ID)
	if err != nil {
		t.Fatal(err)
	}

	posRepo := position.NewGormRepository(db)
	posService := position.NewService(posRepo)

	params := market.DefaultParams()
	pe := market.NewSimplePriceEngine()
	mm := market.NewMarketMaker(params, pe)

	me := NewMatchingEngine(eventService, fundService, posService, mm)

	// Buy 20 NO shares
	order, err := me.BuyNo(u.ID, e.ID, decimal.NewFromInt(20))
	if err != nil {
		t.Fatalf("BuyNo failed: %v", err)
	}

	if order.Status != StatusFilled {
		t.Errorf("expected status FILLED, got %v", order.Status)
	}

	// Check event inventory changed
	eUpdated, err := eventService.GetByID(e.ID)
	if err != nil {
		t.Fatal(err)
	}
	// Original: YesInventory 400 (1000 * (1-0.6)), NoInventory 600 (1000 * 0.6)
	// After buying 20 NO: NoInventory 600 - 20 = 580, YesInventory 400 + 20 = 420
	expectedYesInv := decimal.NewFromInt(400).Add(decimal.NewFromInt(20))
	if !eUpdated.YesInventory.Equal(expectedYesInv) {
		t.Errorf("expected yes inventory %v, got %v", expectedYesInv, eUpdated.YesInventory)
	}
	expectedNoInv := decimal.NewFromInt(600).Sub(decimal.NewFromInt(20))
	if !eUpdated.NoInventory.Equal(expectedNoInv) {
		t.Errorf("expected no inventory %v, got %v", expectedNoInv, eUpdated.NoInventory)
	}

	// Check user position
	userPos, err := posService.GetPosition(u.ID, e.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !userPos.NoQuantity.Equal(decimal.NewFromInt(20)) {
		t.Errorf("expected no quantity 20, got %v", userPos.NoQuantity)
	}
}

func TestSellYes(t *testing.T) {
	db := setupTestDB(t)

	// Setup dependencies
	userRepo := user.NewGormRepository(db)
	userService := user.NewUserService(userRepo)
	u, err := userService.Create("testuser3")
	if err != nil {
		t.Fatal(err)
	}

	fundRepo := fund.NewGormRepository(db)
	fundService := fund.NewService(fundRepo)
	err = fundService.CreateAccount(u)
	if err != nil {
		t.Fatal(err)
	}
	err = fundService.Deposit(u.ID, decimal.NewFromInt(1000))
	if err != nil {
		t.Fatal(err)
	}

	eventRepo := event.NewGormRepository(db)
	eventService := event.NewService(eventRepo)
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now().Add(1 * time.Hour)
	e, err := eventService.Create(
		"Test Event 3",
		"Test Description",
		startTime, endTime,
		decimal.NewFromInt(50).Div(decimal.NewFromInt(100)), // 50% YES
		decimal.NewFromInt(1000), // 1000 shares
	)
	if err != nil {
		t.Fatal(err)
	}
	err = eventService.StartTrading(e.ID)
	if err != nil {
		t.Fatal(err)
	}

	posRepo := position.NewGormRepository(db)
	posService := position.NewService(posRepo)

	params := market.DefaultParams()
	pe := market.NewSimplePriceEngine()
	mm := market.NewMarketMaker(params, pe)

	me := NewMatchingEngine(eventService, fundService, posService, mm)

	// First buy 50 YES
	_, err = me.BuyYes(u.ID, e.ID, decimal.NewFromInt(50))
	if err != nil {
		t.Fatalf("BuyYes failed: %v", err)
	}

	// Get balance before sell
	accBefore, err := fundService.GetAccount(u.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Now sell 20 YES back to platform
	order, err := me.SellYes(u.ID, e.ID, decimal.NewFromInt(20))
	if err != nil {
		t.Fatalf("SellYes failed: %v", err)
	}

	if order.Status != StatusFilled {
		t.Errorf("expected status FILLED, got %v", order.Status)
	}

	// Check user position after sell
	userPos, err := posService.GetPosition(u.ID, e.ID)
	if err != nil {
		t.Fatal(err)
	}
	// Should have 50 - 20 = 30 YES left
	if !userPos.YesQuantity.Equal(decimal.NewFromInt(30)) {
		t.Errorf("expected yes quantity 30, got %v", userPos.YesQuantity)
	}

	// Check balance increased after sell
	accAfter, err := fundService.GetAccount(u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !accAfter.Balance.GreaterThan(accBefore.Balance) {
		t.Errorf("expected balance to increase after selling, went from %v to %v", accBefore.Balance, accAfter.Balance)
	}

	// Check inventory returned to platform
	eUpdated, err := eventService.GetByID(e.ID)
	if err != nil {
		t.Fatal(err)
	}
	// After buying 50: YesInventory 500 - 50 = 450
	// After selling 20: YesInventory 450 + 20 = 470
	expectedYesInv := decimal.NewFromInt(500).Sub(decimal.NewFromInt(50)).Add(decimal.NewFromInt(20))
	if !eUpdated.YesInventory.Equal(expectedYesInv) {
		t.Errorf("expected yes inventory %v, got %v", expectedYesInv, eUpdated.YesInventory)
	}
}

func TestSettleEvent(t *testing.T) {
	db := setupTestDB(t)

	// Setup dependencies
	userRepo := user.NewGormRepository(db)
	userService := user.NewUserService(userRepo)
	u1, err := userService.Create("user1")
	if err != nil {
		t.Fatal(err)
	}
	u2, err := userService.Create("user2")
	if err != nil {
		t.Fatal(err)
	}

	fundRepo := fund.NewGormRepository(db)
	fundService := fund.NewService(fundRepo)
	err = fundService.CreateAccount(u1)
	if err != nil {
		t.Fatal(err)
	}
	err = fundService.CreateAccount(u2)
	if err != nil {
		t.Fatal(err)
	}

	eventRepo := event.NewGormRepository(db)
	eventService := event.NewService(eventRepo)
	startTime := time.Now().Add(-2 * time.Hour)
	endTime := time.Now().Add(-30 * time.Minute)
	e, err := eventService.Create(
		"Settled Event",
		"Event that's ready to settle",
		startTime, endTime,
		decimal.NewFromInt(50).Div(decimal.NewFromInt(100)),
		decimal.NewFromInt(1000),
	)
	if err != nil {
		t.Fatal(err)
	}
	err = eventService.StartTrading(e.ID)
	if err != nil {
		t.Fatal(err)
	}

	posRepo := position.NewGormRepository(db)
	posService := position.NewService(posRepo)

	// User1 has 10 YES
	err = posService.AddPosition(u1.ID, e.ID, decimal.NewFromInt(10), position.DirectionYes)
	if err != nil {
		t.Fatal(err)
	}
	// User2 has 20 NO
	err = posService.AddPosition(u2.ID, e.ID, decimal.NewFromInt(20), position.DirectionNo)
	if err != nil {
		t.Fatal(err)
	}

	// Settle as YES won
	err = eventService.Settle(e.ID, event.ResultYesWon)
	if err != nil {
		t.Fatal(err)
	}

	// Get balances before settlement
	acc1Before, err := fundService.GetAccount(u1.ID)
	if err != nil {
		t.Fatal(err)
	}
	acc2Before, err := fundService.GetAccount(u2.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Execute settlement
	eToSettle, err := eventService.GetByID(e.ID)
	if err != nil {
		t.Fatal(err)
	}
	err = SettleEvent(eToSettle, posService, fundService)
	if err != nil {
		t.Fatalf("SettleEvent failed: %v", err)
	}

	// Check payout: User1 should get 10 units
	acc1After, err := fundService.GetAccount(u1.ID)
	if err != nil {
		t.Fatal(err)
	}
	payout1 := acc1After.Balance.Sub(acc1Before.Balance)
	if !payout1.Equal(decimal.NewFromInt(10)) {
		t.Errorf("expected payout 10 for user1, got %v", payout1)
	}

	// Check user2 gets nothing (NO lost)
	acc2After, err := fundService.GetAccount(u2.ID)
	if err != nil {
		t.Fatal(err)
	}
	payout2 := acc2After.Balance.Sub(acc2Before.Balance)
	if !payout2.Equal(decimal.Zero) {
		t.Errorf("expected payout 0 for user2, got %v", payout2)
	}
}

func TestBuyYesInsufficientFunds(t *testing.T) {
	db := setupTestDB(t)

	// Setup dependencies
	userRepo := user.NewGormRepository(db)
	userService := user.NewUserService(userRepo)
	u, err := userService.Create("pooruser")
	if err != nil {
		t.Fatal(err)
	}

	fundRepo := fund.NewGormRepository(db)
	fundService := fund.NewService(fundRepo)
	err = fundService.CreateAccount(u)
	if err != nil {
		t.Fatal(err)
	}
	// Only deposit 10
	err = fundService.Deposit(u.ID, decimal.NewFromInt(10))
	if err != nil {
		t.Fatal(err)
	}

	eventRepo := event.NewGormRepository(db)
	eventService := event.NewService(eventRepo)
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now().Add(1 * time.Hour)
	e, err := eventService.Create(
		"Test Event",
		"Test Description",
		startTime, endTime,
		decimal.NewFromInt(60).Div(decimal.NewFromInt(100)),
		decimal.NewFromInt(1000),
	)
	if err != nil {
		t.Fatal(err)
	}
	err = eventService.StartTrading(e.ID)
	if err != nil {
		t.Fatal(err)
	}

	posRepo := position.NewGormRepository(db)
	posService := position.NewService(posRepo)

	params := market.DefaultParams()
	pe := market.NewSimplePriceEngine()
	mm := market.NewMarketMaker(params, pe)

	me := NewMatchingEngine(eventService, fundService, posService, mm)

	// Try to buy 100 shares which costs ~60, more than 10 available
	_, err = me.BuyYes(u.ID, e.ID, decimal.NewFromInt(100))
	if err == nil {
		t.Errorf("expected error for insufficient funds, got nil")
	}
}
