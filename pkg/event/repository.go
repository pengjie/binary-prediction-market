package event

import (
	"gorm.io/gorm"
)

// Repository defines the interface for event data access operations.
type Repository interface {
	// Create inserts a new event into the database.
	Create(event *Event) error
	// GetByID retrieves an event by its ID.
	GetByID(id uint) (*Event, error)
	// ListActive returns all events in trading status.
	ListActive() ([]*Event, error)
	// ListAll returns all events.
	ListAll() ([]*Event, error)
	// Update saves changes to an existing event.
	Update(event *Event) error
}

// GormRepository is a GORM-based implementation of Repository.
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new GormRepository instance.
func NewGormRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

// Create inserts a new event into the database.
func (r *GormRepository) Create(event *Event) error {
	return r.db.Create(event).Error
}

// GetByID retrieves an event by its ID.
func (r *GormRepository) GetByID(id uint) (*Event, error) {
	var event Event
	err := r.db.Where("id = ?", id).First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// ListActive returns all events in trading status.
func (r *GormRepository) ListActive() ([]*Event, error) {
	var events []*Event
	err := r.db.Where("status = ?", StatusTrading).Find(&events).Error
	return events, err
}

// ListAll returns all events.
func (r *GormRepository) ListAll() ([]*Event, error) {
	var events []*Event
	err := r.db.Find(&events).Error
	return events, err
}

// Update saves changes to an existing event.
func (r *GormRepository) Update(event *Event) error {
	return r.db.Save(event).Error
}
