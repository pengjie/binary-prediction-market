package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/huinong/golang-claude/pkg/common/response"
	"github.com/huinong/golang-claude/pkg/event"
)

// EventHandler handles event-related API endpoints.
type EventHandler struct {
	service event.Service
}

// NewEventHandler creates a new EventHandler.
func NewEventHandler(s event.Service) *EventHandler {
	return &EventHandler{service: s}
}

// ListEvents lists all actively trading events.
// GET /api/v1/events
func (h *EventHandler) ListEvents(c *gin.Context) {
	events, err := h.service.ListActive()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, http.StatusOK, events)
}

// ListAllEvents lists all events (including pending and settled).
// GET /api/v1/events/all
func (h *EventHandler) ListAllEvents(c *gin.Context) {
	events, err := h.service.ListAll()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, http.StatusOK, events)
}

// GetEvent gets a single event by ID.
// GET /api/v1/events/:id
func (h *EventHandler) GetEvent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, err)
		return
	}
	event, err := h.service.GetByID(uint(id))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, http.StatusOK, event)
}
