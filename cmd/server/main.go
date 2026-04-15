package main

import (
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"

	"github.com/huinong/golang-claude/pkg/api"
	"github.com/huinong/golang-claude/pkg/api/handlers"
	"github.com/huinong/golang-claude/pkg/common/config"
	"github.com/huinong/golang-claude/pkg/event"
	"github.com/huinong/golang-claude/pkg/fund"
	"github.com/huinong/golang-claude/pkg/market"
	"github.com/huinong/golang-claude/pkg/matching"
	"github.com/huinong/golang-claude/pkg/position"
	"github.com/huinong/golang-claude/pkg/user"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Load config
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// Ensure data directory exists
	dataDir := filepath.Dir(cfg.Database.Path)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create data directory: %v", err))
	}

	// Connect database - using modernc.org/sqlite (pure Go, no CGO required)
	// Use driver name "sqlite" for modernc.org/sqlite which registers itself as "sqlite"
	dialector := sqlite.Open(cfg.Database.Path)
	dialector.(*sqlite.Dialector).DriverName = "sqlite"
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %v", err))
	}

	// Auto migrate all database schemas
	db.AutoMigrate(
		&user.User{},
		&fund.Account{},
		&event.Event{},
		&matching.Order{},
		&position.Position{},
	)

	// Initialize repositories
	userRepo := user.NewGormRepository(db)
	fundRepo := fund.NewGormRepository(db)
	eventRepo := event.NewGormRepository(db)
	positionRepo := position.NewGormRepository(db)

	// Initialize services
	userService := user.NewUserService(userRepo)
	fundService := fund.NewService(fundRepo)
	eventService := event.NewService(eventRepo)
	positionService := position.NewService(positionRepo)

	// Initialize market maker
	params := market.DefaultParams()
	priceEngine := market.NewSimplePriceEngine()
	marketMaker := market.NewMarketMaker(params, priceEngine)

	// Initialize matching engine
	matchingEngine := matching.NewMatchingEngine(eventService, fundService, positionService, marketMaker)

	// Initialize handlers
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

	// Start server
	fmt.Printf("Server starting on port %d...\n", cfg.Server.Port)
	panic(router.Run(fmt.Sprintf(":%d", cfg.Server.Port)))
}
