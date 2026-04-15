package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"github.com/huinong/golang-claude/pkg/common/decimal"
	"github.com/huinong/golang-claude/pkg/event"
	"github.com/huinong/golang-claude/pkg/fund"
	"github.com/huinong/golang-claude/pkg/market"
	"github.com/huinong/golang-claude/pkg/matching"
	"github.com/huinong/golang-claude/pkg/position"
	"github.com/huinong/golang-claude/pkg/user"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Backtest is a command-line tool for backtesting the three-force market making algorithm
// with historical simulation of multiple trading scenarios.

func main() {
	// Parse command line flags
	dbPath := flag.String("db", "./data/backtest.db", "Path to backtest database file")
	flag.Parse()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(*dbPath), 0755); err != nil {
		log.Fatalf("failed to create data directory: %v", err)
	}

	// Open database - using modernc.org/sqlite (pure Go, no CGO required)
	// Use driver name "sqlite" for modernc.org/sqlite which registers itself as "sqlite"
	dialector := sqlite.Open(*dbPath)
	dialector.(*sqlite.Dialector).DriverName = "sqlite"
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Auto migrate
	db.AutoMigrate(
		&user.User{},
		&fund.Account{},
		&event.Event{},
		&matching.Order{},
		&position.Position{},
	)

	// Initialize dependencies
	userRepo := user.NewGormRepository(db)
	fundRepo := fund.NewGormRepository(db)
	eventRepo := event.NewGormRepository(db)
	positionRepo := position.NewGormRepository(db)

	userService := user.NewUserService(userRepo)
	fundService := fund.NewService(fundRepo)
	eventService := event.NewService(eventRepo)
	positionService := position.NewService(positionRepo)

	// Create market maker with default parameters
	params := market.DefaultParams()
	priceEngine := market.NewSimplePriceEngine()
	marketMaker := market.NewMarketMaker(params, priceEngine)

	// Create matching engine
	matchingEngine := matching.NewMatchingEngine(eventService, fundService, positionService, marketMaker)

	// Create test event
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now().Add(24 * time.Hour)

	e, err := eventService.Create(
		"Test Backtest Event",
		"Simulation event for backtesting the three-force model",
		startTime,
		endTime,
		decimal.NewFromInt(60).Div(decimal.NewFromInt(100)), // 60% probability
		decimal.NewFromInt(1000), // 1000 initial supply
	)
	if err != nil {
		log.Fatalf("failed to create event: %v", err)
	}

	err = eventService.StartTrading(e.ID)
	if err != nil {
		log.Fatalf("failed to start trading: %v", err)
	}

	// Create test user
	testUser, err := userService.Create("backtest-user")
	if err != nil {
		log.Fatalf("failed to create user: %v", err)
	}

	err = fundService.CreateAccount(testUser)
	if err != nil {
		log.Fatalf("failed to create account: %v", err)
	}

	// Deposit initial capital
	err = fundService.Deposit(testUser.ID, decimal.NewFromInt(10000))
	if err != nil {
		log.Fatalf("failed to deposit: %v", err)
	}

	fmt.Printf("Backtest setup complete:\n")
	fmt.Printf("  Event ID: %d\n", e.ID)
	fmt.Printf("  Initial YES Price: %.4f\n", e.YesPrice.InexactFloat64())
	fmt.Printf("  Initial YES Inventory: %.0f\n", e.YesInventory.InexactFloat64())
	fmt.Printf("  Initial NO Inventory: %.0f\n", e.NoInventory.InexactFloat64())
	fmt.Printf("  User balance: 10000\n\n")

	// Print market parameters
	fmt.Printf("Market maker parameters:\n")
	fmt.Printf("  Base Spread: %.4f\n", params.BaseSpread.InexactFloat64())
	fmt.Printf("  Alpha (inventory coefficient): %.2f\n", params.Alpha.InexactFloat64())
	fmt.Printf("  Beta (probability coefficient): %.2f\n", params.Beta.InexactFloat64())
	fmt.Printf("  Gamma (liquidity coefficient): %.2f\n", params.Gamma.InexactFloat64())
	fmt.Println()

	// Calculate initial quotation
	quote := marketMaker.CalculateQuotation(e, decimal.NewFromInt(1))
	fmt.Printf("Initial quotation:\n")
	fmt.Printf("  YES Bid: %.4f, YES Ask: %.4f (spread: %.4f)\n",
		quote.YesBid.InexactFloat64(), quote.YesAsk.InexactFloat64(),
		quote.YesAsk.Sub(quote.YesBid).InexactFloat64())
	fmt.Printf("  NO Bid: %.4f, NO Ask: %.4f (spread: %.4f)\n",
		quote.NoBid.InexactFloat64(), quote.NoAsk.InexactFloat64(),
		quote.NoAsk.Sub(quote.NoBid).InexactFloat64())
	fmt.Println()

	// Run a simulation of a few trades
	trades := []struct {
		direction string
		quantity  float64
	}{
		{"YES", 100},
		{"NO", 50},
		{"YES", -30}, // Sell 30 YES
	}

	for i, trade := range trades {
		fmt.Printf("Trade %d: %s %.0f shares\n", i+1, trade.direction, trade.quantity)

		var order *matching.Order
		qty := decimal.NewFromFloat(trade.quantity)

		switch {
		case trade.direction == "YES" && trade.quantity > 0:
			order, err = matchingEngine.BuyYes(testUser.ID, e.ID, qty)
		case trade.direction == "YES" && trade.quantity < 0:
			order, err = matchingEngine.SellYes(testUser.ID, e.ID, qty.Abs())
		case trade.direction == "NO" && trade.quantity > 0:
			order, err = matchingEngine.BuyNo(testUser.ID, e.ID, qty)
		case trade.direction == "NO" && trade.quantity < 0:
			order, err = matchingEngine.SellNo(testUser.ID, e.ID, qty.Abs())
		}

		if err != nil {
			log.Printf("  ERROR: %v", err)
			continue
		}

		// Get updated event
		e, _ = eventService.GetByID(e.ID)
		newQuote := marketMaker.CalculateQuotation(e, decimal.NewFromInt(1))

		fmt.Printf("  Filled at price: %.4f, total amount: %.4f\n",
			order.Price.InexactFloat64(), order.Amount.InexactFloat64())
		fmt.Printf("  New YES Price: %.4f, YES Inventory: %.0f\n",
			e.YesPrice.InexactFloat64(), e.YesInventory.InexactFloat64())
		fmt.Printf("  New quotation YES: %.4f - %.4f\n",
			newQuote.YesBid.InexactFloat64(), newQuote.YesAsk.InexactFloat64())
		fmt.Println()
	}

	// Get final user balance
	account, err := fundService.GetAccount(testUser.ID)
	if err != nil {
		log.Fatalf("failed to get account: %v", err)
	}

	fmt.Printf("Backtest complete:\n")
	fmt.Printf("  Final user balance: %.2f\n", account.Balance.InexactFloat64())
	fmt.Printf("  Final YES Price: %.4f\n", e.YesPrice.InexactFloat64())
	fmt.Printf("  PnL: %.2f\n", account.Balance.Sub(decimal.NewFromInt(10000)).InexactFloat64())
}
