package user

import "gorm.io/gorm"

type Repository interface {
	Create(user *User) error
	GetByID(id uint) (*User, error)
	GetByAPIKey(apiKey string) (*User, error)
	GetByUsername(username string) (*User, error)
	Update(user *User) error
	Delete(id uint) error
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Create(user *User) error {
	return r.db.Create(user).Error
}

func (r *GormRepository) GetByID(id uint) (*User, error) {
	var u User
	err := r.db.Where("id = ?", id).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *GormRepository) GetByAPIKey(apiKey string) (*User, error) {
	var u User
	err := r.db.Where("api_key = ?", apiKey).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *GormRepository) GetByUsername(username string) (*User, error) {
	var u User
	err := r.db.Where("username = ?", username).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *GormRepository) Update(user *User) error {
	return r.db.Save(user).Error
}

func (r *GormRepository) Delete(id uint) error {
	return r.db.Delete(&User{}, id).Error
}
