package api

import (
	"github.com/gin-gonic/gin"
	"github.com/huinong/golang-claude/pkg/api/handlers"
)

// SetupRouter configures all routes and returns the gin.Engine.
func SetupRouter(
	eventHandler *handlers.EventHandler,
	userHandler *handlers.UserHandler,
	tradeHandler *handlers.TradeHandler,
	adminHandler *handlers.AdminHandler,
	authMiddleware gin.HandlerFunc,
) *gin.Engine {
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Public routes - no authentication required
	public := r.Group("/api/v1")
	{
		public.GET("/events", eventHandler.ListEvents)
		public.GET("/events/all", eventHandler.ListAllEvents)
		public.GET("/events/:id", eventHandler.GetEvent)
		public.POST("/users", userHandler.CreateUser)
	}

	// Authenticated routes - require valid X-API-Key
	auth := r.Group("/api/v1")
	auth.Use(authMiddleware)
	{
		// User account endpoints
		auth.GET("/user/account", userHandler.GetAccount)
		auth.POST("/user/deposit", userHandler.Deposit)

		// Trading endpoints
		auth.POST("/trade/buy", tradeHandler.Buy)
		auth.POST("/trade/sell", tradeHandler.Sell)
	}

	// Admin endpoints - for now, same authentication (in production would need admin flag)
	admin := r.Group("/api/v1/admin")
	admin.Use(authMiddleware)
	{
		admin.POST("/events", adminHandler.CreateEvent)
		admin.POST("/events/:id/start", adminHandler.StartTrading)
		admin.POST("/events/:id/settle", adminHandler.SettleEvent)
	}

	return r
}
