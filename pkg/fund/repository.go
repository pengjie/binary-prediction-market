package fund

import (
	"gorm.io/gorm"
)

// Repository defines the interface for account data access operations.
type Repository interface {
	// Create inserts a new account into the database.
	Create(account *Account) error
	// GetByUserID finds an account by user ID.
	GetByUserID(userID uint) (*Account, error)
	// Update saves changes to an existing account.
	Update(account *Account) error
}

// GormRepository is a GORM-based implementation of Repository.
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new GormRepository instance.
func NewGormRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

// Create inserts a new account into the database.
func (r *GormRepository) Create(account *Account) error {
	return r.db.Create(account).Error
}

// GetByUserID finds an account by user ID.
func (r *GormRepository) GetByUserID(userID uint) (*Account, error) {
	var account Account
	err := r.db.Where("user_id = ?", userID).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// Update saves changes to an existing account.
func (r *GormRepository) Update(account *Account) error {
	return r.db.Save(account).Error
}
