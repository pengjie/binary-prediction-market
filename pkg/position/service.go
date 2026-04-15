package position

import (
	"github.com/huinong/golang-claude/pkg/common/decimal"
	"github.com/huinong/golang-claude/pkg/common/errors"
	"gorm.io/gorm"
)

// Direction indicates whether we're adding YES or NO shares.
type Direction int

const (
	DirectionYes = Direction(1)
	DirectionNo  = Direction(2)
)

// Service defines the interface for position management operations.
type Service interface {
	// AddPosition adds a new position (shares) to a user's holdings for an event.
	// If no position exists, creates a new one.
	AddPosition(userID uint, eventID uint, quantity decimal.Decimal, direction Direction) error
	// GetPosition gets the current position for a user and event.
	GetPosition(userID uint, eventID uint) (*Position, error)
	// ListByEvent lists all positions for an event (needed for settlement).
	ListByEvent(eventID uint) ([]*Position, error)
	// ListByUserID lists all positions for a user.
	ListByUserID(userID uint) ([]*Position, error)
}

type service struct {
	repo Repository
}

// NewService creates a new position service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// AddPosition adds shares to an existing position or creates a new one.
func (s *service) AddPosition(userID uint, eventID uint, quantity decimal.Decimal, direction Direction) error {
	pos, err := s.repo.GetByUserIDAndEventID(userID, eventID)
	if err != nil {
		// Not found, create new position
		if err == gorm.ErrRecordNotFound {
			pos = &Position{
				UserID:      userID,
				EventID:     eventID,
				YesQuantity: decimal.Zero,
				NoQuantity:  decimal.Zero,
			}
			switch direction {
			case DirectionYes:
				pos.YesQuantity = quantity
			case DirectionNo:
				pos.NoQuantity = quantity
			}
			return s.repo.Create(pos)
		}
		return err
	}

	// Found existing position, add to it
	switch direction {
	case DirectionYes:
		pos.YesQuantity = pos.YesQuantity.Add(quantity)
	case DirectionNo:
		pos.NoQuantity = pos.NoQuantity.Add(quantity)
	}

	return s.repo.Update(pos)
}

// GetPosition gets the current position for a user and event.
func (s *service) GetPosition(userID uint, eventID uint) (*Position, error) {
	pos, err := s.repo.GetByUserIDAndEventID(userID, eventID)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "position not found")
	}
	return pos, nil
}

// ListByEvent lists all positions for an event.
func (s *service) ListByEvent(eventID uint) ([]*Position, error) {
	return s.repo.ListByEvent(eventID)
}

// ListByUserID lists all positions for a user.
func (s *service) ListByUserID(userID uint) ([]*Position, error) {
	return s.repo.ListByUserID(userID)
}
