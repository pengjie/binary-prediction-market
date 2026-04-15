package event

import (
	"time"

	"github.com/huinong/golang-claude/pkg/common/decimal"
	"github.com/huinong/golang-claude/pkg/common/errors"
)

// Service defines the interface for event management operations.
type Service interface {
	// Create creates a new binary event with initial inventory allocation.
	Create(title, description string, startTime, endTime time.Time, initialYesPrice decimal.Decimal, initialSupply decimal.Decimal) (*Event, error)
	// GetByID retrieves an event by ID.
	GetByID(id uint) (*Event, error)
	// ListActive returns all actively trading events.
	ListActive() ([]*Event, error)
	// ListAll returns all events.
	ListAll() ([]*Event, error)
	// StartTrading changes event status from PENDING to TRADING.
	StartTrading(id uint) error
	// Settle changes event status from TRADING to SETTLED with the final result.
	Settle(id uint, result Result) error
	// UpdatePrice updates the current YES price of the event (called by market maker).
	UpdatePrice(id uint, newYesPrice decimal.Decimal) error
}

type service struct {
	repo Repository
}

// NewService creates a new event service instance.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// Create creates a new binary event with initial inventory allocation.
// The initial inventory is allocated based on the initial price:
// - YesInventory = initialSupply * (1 - initialYesPrice)
// - NoInventory = initialSupply * initialYesPrice
func (s *service) Create(
	title, description string,
	startTime, endTime time.Time,
	initialYesPrice decimal.Decimal,
	initialSupply decimal.Decimal,
) (*Event, error) {
	// Validate inputs
	if title == "" {
		return nil, errors.New(errors.CodeInvalidInput, "title cannot be empty")
	}
	if len(title) > 200 {
		return nil, errors.New(errors.CodeInvalidInput, "title cannot exceed 200 characters")
	}
	if !endTime.After(startTime) {
		return nil, errors.New(errors.CodeInvalidInput, "end time must be after start time")
	}
	if initialYesPrice.LessThan(decimal.NewFromInt(1).Div(decimal.NewFromInt(100))) ||
		initialYesPrice.GreaterThan(decimal.NewFromInt(99).Div(decimal.NewFromInt(100))) {
		return nil, errors.New(errors.CodeInvalidInput, "initial price must be between 0.01 and 0.99")
	}
	if initialSupply.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New(errors.CodeInvalidInput, "initial supply must be positive")
	}

	event := &Event{
		Title:         title,
		Description:   description,
		StartTime:     startTime,
		EndTime:       endTime,
		Status:        StatusPending,
		YesPrice:      initialYesPrice,
		YesInventory:  initialSupply.Mul(decimal.One.Sub(initialYesPrice)),
		NoInventory:   initialSupply.Mul(initialYesPrice),
	}
	err := s.repo.Create(event)
	if err != nil {
		return nil, err
	}
	return event, nil
}

// GetByID retrieves an event by ID.
func (s *service) GetByID(id uint) (*Event, error) {
	return s.repo.GetByID(id)
}

// ListActive returns all actively trading events.
func (s *service) ListActive() ([]*Event, error) {
	return s.repo.ListActive()
}

// ListAll returns all events.
func (s *service) ListAll() ([]*Event, error) {
	return s.repo.ListAll()
}

// StartTrading changes event status from PENDING to TRADING.
func (s *service) StartTrading(id uint) error {
	event, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New(errors.CodeNotFound, "event not found")
	}
	if event.Status != StatusPending {
		return errors.New(errors.CodeInvalidInput, "event already started")
	}
	event.Status = StatusTrading
	return s.repo.Update(event)
}

// Settle changes event status from TRADING to SETTLED with the final result.
func (s *service) Settle(id uint, result Result) error {
	event, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New(errors.CodeNotFound, "event not found")
	}
	if event.Status != StatusTrading {
		return errors.New(errors.CodeInvalidInput, "event not in trading status")
	}
	event.Status = StatusSettled
	event.Result = result
	return s.repo.Update(event)
}

// UpdatePrice updates the current YES price of the event.
func (s *service) UpdatePrice(id uint, newYesPrice decimal.Decimal) error {
	event, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New(errors.CodeNotFound, "event not found")
	}
	// Clamp price to valid range
	if newYesPrice.LessThan(decimal.NewFromInt(1).Div(decimal.NewFromInt(100))) {
		newYesPrice = decimal.NewFromInt(1).Div(decimal.NewFromInt(100))
	}
	if newYesPrice.GreaterThan(decimal.NewFromInt(99).Div(decimal.NewFromInt(100))) {
		newYesPrice = decimal.NewFromInt(99).Div(decimal.NewFromInt(100))
	}
	event.YesPrice = newYesPrice
	return s.repo.Update(event)
}
