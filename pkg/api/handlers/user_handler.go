package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huinong/golang-claude/pkg/common/decimal"
	"github.com/huinong/golang-claude/pkg/common/errors"
	"github.com/huinong/golang-claude/pkg/common/response"
	"github.com/huinong/golang-claude/pkg/fund"
	"github.com/huinong/golang-claude/pkg/user"
)

// UserHandler handles user and account-related API endpoints.
type UserHandler struct {
	userService user.Service
	fundService fund.Service
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(us user.Service, fs fund.Service) *UserHandler {
	return &UserHandler{userService: us, fundService: fs}
}

// GetAccount gets the user's current fund account information.
// GET /api/v1/user/account
func (h *UserHandler) GetAccount(c *gin.Context) {
	userID := user.GetUserID(c)
	if userID == 0 {
		response.Error(c, errors.New(errors.CodeUnauthorized, "unauthorized"))
		return
	}
	account, err := h.fundService.GetAccount(userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, http.StatusOK, account)
}

// DepositRequest represents a deposit request body.
type DepositRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

// Deposit deposits funds to the user's account.
// POST /api/v1/user/deposit
func (h *UserHandler) Deposit(c *gin.Context) {
	userID := user.GetUserID(c)
	if userID == 0 {
		response.Error(c, errors.New(errors.CodeUnauthorized, "unauthorized"))
		return
	}

	var req DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.CodeBadRequest, err.Error()))
		return
	}

	amount := decimal.NewFromFloat(req.Amount)
	err := h.fundService.Deposit(userID, amount)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, http.StatusOK, nil)
}

// CreateUserRequest represents a create user request body.
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
}

// CreateUser creates a new user.
// POST /api/v1/user/create
// This is open for public access in MVP to allow user creation without admin approval
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.CodeBadRequest, err.Error()))
		return
	}

	user, err := h.userService.Create(req.Username)
	if err != nil {
		response.Error(c, err)
		return
	}

	// Create fund account
	err = h.fundService.CreateAccount(user)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, http.StatusCreated, user)
}
