package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/huinong/golang-claude/pkg/common/decimal"
	"github.com/huinong/golang-claude/pkg/common/errors"
	"github.com/huinong/golang-claude/pkg/common/response"
	"github.com/huinong/golang-claude/pkg/event"
)

// AdminHandler handles admin endpoints for event management.
type AdminHandler struct {
	eventService event.Service
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(es event.Service) *AdminHandler {
	return &AdminHandler{eventService: es}
}

// CreateEventRequest represents the request to create a new event.
type CreateEventRequest struct {
	Title         string    `json:"title" binding:"required,min=1,max=200"`
	Description   string    `json:"description" binding:"max=1000"`
	StartTime     time.Time `json:"start_time" binding:"required"`
	EndTime       time.Time `json:"end_time" binding:"required"`
	InitialYesPrice float64 `json:"initial_yes_price" binding:"required,gt=0,lt=1"`
	InitialSupply  float64 `json:"initial_supply" binding:"required,gt=0"`
}

// CreateEvent creates a new binary event.
// POST /api/v1/admin/events
func (h *AdminHandler) CreateEvent(c *gin.Context) {
	var req CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.CodeBadRequest, err.Error()))
		return
	}

	if req.StartTime.After(req.EndTime) {
		response.Error(c, errors.New(errors.CodeBadRequest, "start time must be before end time"))
		return
	}

	initialYesPrice := decimal.NewFromFloat(req.InitialYesPrice)
	initialSupply := decimal.NewFromFloat(req.InitialSupply)

	event, err := h.eventService.Create(
		req.Title,
		req.Description,
		req.StartTime,
		req.EndTime,
		initialYesPrice,
		initialSupply,
	)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, http.StatusCreated, event)
}

// StartTrading starts trading for an event.
// POST /api/v1/admin/events/:id/start
func (h *AdminHandler) StartTrading(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, errors.New(errors.CodeBadRequest, "invalid event ID"))
		return
	}

	err = h.eventService.StartTrading(uint(id))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, http.StatusOK, gin.H{"message": "trading started"})
}

// SettleEventRequest represents the request to settle an event.
type SettleEventRequest struct {
	Result string `json:"result" binding:"required,oneof=YES NO"`
}

// SettleEvent settles an event with the final result and distributes payouts.
// POST /api/v1/admin/events/:id/settle
func (h *AdminHandler) SettleEvent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, errors.New(errors.CodeBadRequest, "invalid event ID"))
		return
	}

	var req SettleEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.CodeBadRequest, err.Error()))
		return
	}

	var result event.Result
	if req.Result == "YES" {
		result = event.ResultYesWon
	} else {
		result = event.ResultNoWon
	}

	err = h.eventService.Settle(uint(id), result)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, http.StatusOK, gin.H{"message": "event settled"})
}
