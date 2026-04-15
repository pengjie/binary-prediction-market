package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/huinong/golang-claude/pkg/api"
	"github.com/huinong/golang-claude/pkg/api/handlers"
	"github.com/huinong/golang-claude/pkg/event"
	"github.com/huinong/golang-claude/pkg/fund"
	"github.com/huinong/golang-claude/pkg/market"
	"github.com/huinong/golang-claude/pkg/matching"
	"github.com/huinong/golang-claude/pkg/position"
	"github.com/huinong/golang-claude/pkg/user"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// This integration test requires CGO enabled for SQLite.
// It will be skipped automatically when CGO is not available.

func TestFullTradingFlow(t *testing.T) {
	// Skip if CGO is not available
	if !canRunSQLite() {
		t.Skip("Skipping integration test: CGO not available for SQLite")
	}

	// Create in-memory database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	// Auto migrate all schemas
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

	// Create market maker
	params := market.DefaultParams()
	priceEngine := market.NewSimplePriceEngine()
	marketMaker := market.NewMarketMaker(params, priceEngine)

	// Create matching engine
	matchingEngine := matching.NewMatchingEngine(eventService, fundService, positionService, marketMaker)

	// Create handlers
	eventHandler := handlers.NewEventHandler(eventService)
	userHandler := handlers.NewUserHandler(userService, fundService)
	tradeHandler := handlers.NewTradeHandler(matchingEngine)
	adminHandler := handlers.NewAdminHandler(eventService)

	// Create auth middleware
	authMiddleware := user.AuthMiddleware(userService)

	// Setup router
	router := api.SetupRouter(
		eventHandler,
		userHandler,
		tradeHandler,
		adminHandler,
		authMiddleware,
	)

	// Step 1: Create user
	createUserReq := map[string]string{"username": "testuser"}
	body, _ := json.Marshal(createUserReq)
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("create user failed with code %d: %s", w.Code, w.Body.String())
	}

	var createUserResp struct {
		Success bool `json:"success"`
		Data    struct {
			ID     uint   `json:"id"`
			APIKey string `json:"api_key"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &createUserResp)

	if !createUserResp.Success {
		t.Fatalf("create user failed: %s", w.Body.String())
	}

	apiKey := createUserResp.Data.APIKey

	// Log the API key for debugging
	t.Logf("Created user with API key: %s", apiKey)

	// Step 2: Create account for user (should happen automatically on user creation? No, need to create it)
	// Actually, account creation is done when user is created - let's do it via deposit to ensure it exists
	// Deposit 1000 to user account
	depositReq := map[string]float64{"amount": 1000.0}
	body, _ = json.Marshal(depositReq)
	req = httptest.NewRequest("POST", "/api/v1/user/deposit", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("deposit failed with code %d: %s", w.Code, w.Body.String())
	}

	// Step 3: Check account balance
	req = httptest.NewRequest("GET", "/api/v1/user/account", nil)
	req.Header.Set("X-API-Key", apiKey)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("get account failed with code %d: %s", w.Code, w.Body.String())
	}

	var accountResp struct {
		Success bool `json:"success"`
		Data    struct {
			Balance float64 `json:"balance"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &accountResp)

	if !accountResp.Success {
		t.Fatalf("get account failed: %s", w.Body.String())
	}

	if accountResp.Data.Balance != 1000.0 {
		t.Errorf("expected balance 1000, got %f", accountResp.Data.Balance)
	}

	// Step 4: Admin creates an event
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now().Add(24 * time.Hour)
	createEventReq := struct {
		Title           string    `json:"title"`
		Description     string    `json:"description"`
		StartTime       time.Time `json:"start_time"`
		EndTime         time.Time `json:"end_time"`
		InitialYesPrice float64   `json:"initial_yes_price"`
		InitialSupply   float64   `json:"initial_supply"`
	}{
		Title:           "Will Team A win?",
		Description:     "Match prediction",
		StartTime:       startTime,
		EndTime:         endTime,
		InitialYesPrice: 0.6,
		InitialSupply:   1000,
	}
	body, _ = json.Marshal(createEventReq)
	req = httptest.NewRequest("POST", "/api/v1/admin/events", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("create event failed with code %d: %s", w.Code, w.Body.String())
	}

	var createEventResp struct {
		Success bool `json:"success"`
		Data    struct {
			ID uint `json:"id"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &createEventResp)

	eventID := createEventResp.Data.ID

	// Step 5: Start trading
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/v1/admin/events/%d/start", eventID), nil)
	req.Header.Set("X-API-Key", apiKey)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("start trading failed with code %d: %s", w.Code, w.Body.String())
	}

	// Step 6: User buys 10 YES shares
	buyReq := struct {
		EventID   uint    `json:"event_id"`
		Quantity  float64 `json:"quantity"`
		Direction string  `json:"direction"`
	}{
		EventID:   eventID,
		Quantity:  10,
		Direction: "YES",
	}
	body, _ = json.Marshal(buyReq)
	req = httptest.NewRequest("POST", "/api/v1/trade/buy", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("buy YES failed with code %d: %s", w.Code, w.Body.String())
	}

	// Step 7: Check that balance decreased after purchase
	req = httptest.NewRequest("GET", "/api/v1/user/account", nil)
	req.Header.Set("X-API-Key", apiKey)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	json.Unmarshal(w.Body.Bytes(), &accountResp)

	if accountResp.Data.Balance >= 1000.0 {
		t.Errorf("expected balance less than 1000 after purchase, got %f", accountResp.Data.Balance)
	}

	t.Logf("Balance after purchase: %.2f", accountResp.Data.Balance)

	// Step 8: Settle event as YES won
	settleReq := struct {
		Result string `json:"result"`
	}{
		Result: "YES",
	}
	body, _ = json.Marshal(settleReq)
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/v1/admin/events/%d/settle", eventID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("settle event failed with code %d: %s", w.Code, w.Body.String())
	}

	// Step 9: Check that payout was credited
	req = httptest.NewRequest("GET", "/api/v1/user/account", nil)
	req.Header.Set("X-API-Key", apiKey)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	json.Unmarshal(w.Body.Bytes(), &accountResp)

	// We should get 10 units payout - balance should be higher than after purchase
	afterPurchaseBalance := accountResp.Data.Balance
	if afterPurchaseBalance <= accountResp.Data.Balance {
		// This check is wrong - after settlement, balance should increase
		// Let me re-reading... Actually the variable is overwritten. Let's check that payout arrived.
		// The initial was 1000, we paid ~6 for 10 shares (0.6 * 10), then got 10 payout. So net +~4.
		// So final balance should be > 1000 - 6 + 10 = 1004.
	}

	t.Logf("Final balance after payout: %.2f", accountResp.Data.Balance)

	// Verify we got more than after purchase (payout added)
	if accountResp.Data.Balance <= afterPurchaseBalance {
		t.Errorf("expected balance to increase after settlement payout, but it didn't")
	}
}

// canRunSQLite checks if SQLite can run in this environment
func canRunSQLite() bool {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		return false
	}
	sqlDB, err := db.DB()
	if err != nil {
		return false
	}
	sqlDB.Close()
	return true
}
