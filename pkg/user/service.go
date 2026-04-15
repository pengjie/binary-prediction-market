package user

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/huinong/golang-claude/pkg/common/errors"
)

type Service interface {
	Create(username string) (*User, error)
	GetByID(id uint) (*User, error)
	GetByAPIKey(apiKey string) (*User, error)
	GetByUsername(username string) (*User, error)
	Update(user *User) error
	Delete(id uint) error
}

type service struct {
	repo Repository
}

func NewUserService(repo Repository) Service {
	return &service{repo: repo}
}

// generateAPIKey generates a random 32-byte API key encoded as hex.
func generateAPIKey() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *service) Create(username string) (*User, error) {
	user := &User{
		Username: username,
		APIKey:   generateAPIKey(),
	}
	err := s.repo.Create(user)
	if err != nil {
		return nil, errors.New(errors.CodeInternalError, "failed to create user")
	}
	return user, nil
}

func (s *service) GetByID(id uint) (*User, error) {
	return s.repo.GetByID(id)
}

func (s *service) GetByAPIKey(apiKey string) (*User, error) {
	return s.repo.GetByAPIKey(apiKey)
}

func (s *service) GetByUsername(username string) (*User, error) {
	return s.repo.GetByUsername(username)
}

func (s *service) Update(user *User) error {
	return s.repo.Update(user)
}

func (s *service) Delete(id uint) error {
	return s.repo.Delete(id)
}
