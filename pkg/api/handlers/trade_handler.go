package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huinong/golang-claude/pkg/common/decimal"
	"github.com/huinong/golang-claude/pkg/common/errors"
	"github.com/huinong/golang-claude/pkg/common/response"
	"github.com/huinong/golang-claude/pkg/matching"
	"github.com/huinong/golang-claude/pkg/user"
)

// TradeHandler handles trading-related API endpoints (buy/sell).
type TradeHandler struct {
	matchingEngine *matching.MatchingEngine
}

// NewTradeHandler creates a new TradeHandler.
func NewTradeHandler(me *matching.MatchingEngine) *TradeHandler {
	return &TradeHandler{matchingEngine: me}
}

// BuyRequest represents a buy order request.
type BuyRequest struct {
	EventID  uint    `json:"event_id" binding:"required"`
	Quantity float64 `json:"quantity" binding:"required,gt=0"`
	Direction string `json:"direction" binding:"required,oneof=YES NO"`
}

// Buy handles buying shares from the market maker.
// POST /api/v1/trade/buy
func (h *TradeHandler) Buy(c *gin.Context) {
	userID := user.GetUserID(c)
	if userID == 0 {
		response.Error(c, errors.New(errors.CodeUnauthorized, "unauthorized"))
		return
	}

	var req BuyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.CodeBadRequest, err.Error()))
		return
	}

	qty := decimal.NewFromFloat(req.Quantity)
	var order *matching.Order
	var err error

	if req.Direction == "YES" {
		order, err = h.matchingEngine.BuyYes(userID, req.EventID, qty)
	} else {
		order, err = h.matchingEngine.BuyNo(userID, req.EventID, qty)
	}

	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, http.StatusOK, order)
}

// SellRequest represents a sell order request.
type SellRequest struct {
	EventID  uint    `json:"event_id" binding:"required"`
	Quantity float64 `json:"quantity" binding:"required,gt=0"`
	Direction string `json:"direction" binding:"required,oneof=YES NO"`
}

// Sell handles selling shares back to the market maker.
// POST /api/v1/trade/sell
func (h *TradeHandler) Sell(c *gin.Context) {
	userID := user.GetUserID(c)
	if userID == 0 {
		response.Error(c, errors.New(errors.CodeUnauthorized, "unauthorized"))
		return
	}

	var req SellRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.CodeBadRequest, err.Error()))
		return
	}

	qty := decimal.NewFromFloat(req.Quantity)
	var order *matching.Order
	var err error

	if req.Direction == "YES" {
		order, err = h.matchingEngine.SellYes(userID, req.EventID, qty)
	} else {
		order, err = h.matchingEngine.SellNo(userID, req.EventID, qty)
	}

	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, http.StatusOK, order)
}

// GetQuotationRequest gets a quotation for a given quantity before trading.
type GetQuotationRequest struct {
	EventID uint   `uri:"event_id" binding:"required"`
	Direction string `form:"direction" binding:"required,oneof=YES NO"`
	Quantity float64 `form:"quantity" binding:"required,gt=0"`
}
