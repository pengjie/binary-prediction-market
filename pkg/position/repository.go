package position

import (
	"gorm.io/gorm"
)

// Repository defines the interface for position data access.
type Repository interface {
	Create(pos *Position) error
	GetByUserIDAndEventID(userID uint, eventID uint) (*Position, error)
	ListByEvent(eventID uint) ([]*Position, error)
	ListByUserID(userID uint) ([]*Position, error)
	Update(pos *Position) error
}

// GormRepository is a GORM-based implementation of Repository.
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new GormRepository.
func NewGormRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

// Create creates a new position.
func (r *GormRepository) Create(pos *Position) error {
	return r.db.Create(pos).Error
}

// GetByUserIDAndEventID gets a position by user ID and event ID.
func (r *GormRepository) GetByUserIDAndEventID(userID uint, eventID uint) (*Position, error) {
	var pos Position
	err := r.db.Where("user_id = ? AND event_id = ?", userID, eventID).First(&pos).Error
	if err != nil {
		return nil, err
	}
	return &pos, nil
}

// ListByEvent lists all positions for an event.
func (r *GormRepository) ListByEvent(eventID uint) ([]*Position, error) {
	var positions []*Position
	err := r.db.Where("event_id = ?", eventID).Find(&positions).Error
	return positions, err
}

// ListByUserID lists all positions for a user.
func (r *GormRepository) ListByUserID(userID uint) ([]*Position, error) {
	var positions []*Position
	err := r.db.Where("user_id = ?", userID).Find(&positions).Error
	return positions, err
}

// Update updates an existing position.
func (r *GormRepository) Update(pos *Position) error {
	return r.db.Save(pos).Error
}
