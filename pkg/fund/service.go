package fund

import (
	"github.com/huinong/golang-claude/pkg/common/errors"
	"github.com/huinong/golang-claude/pkg/common/decimal"
	"github.com/huinong/golang-claude/pkg/user"
)

// Service defines the interface for fund account operations.
type Service interface {
	// CreateAccount creates a new account for a user.
	CreateAccount(user *user.User) error
	// GetAccount retrieves an account by user ID.
	GetAccount(userID uint) (*Account, error)
	// Deposit adds funds to a user's account.
	Deposit(userID uint, amount decimal.Decimal) error
	// Withdraw removes funds from a user's account.
	Withdraw(userID uint, amount decimal.Decimal) error
	// Lock reserves funds for an active trade (deducts from available balance).
	Lock(userID uint, amount decimal.Decimal) error
	// Unlock returns reserved funds back to available balance after trade settlement.
	Unlock(userID uint, amount decimal.Decimal) error
	// AddProfit adds profit from a settled trade to the user's balance.
	AddProfit(userID uint, amount decimal.Decimal) error
}

type service struct {
	repo Repository
}

// NewService creates a new fund service instance.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// CreateAccount creates a new account for a user with zero initial balance.
func (s *service) CreateAccount(user *user.User) error {
	account := &Account{
		UserID:        user.ID,
		Balance:       decimal.Zero,
		TotalDeposit:  decimal.Zero,
		TotalWithdraw: decimal.Zero,
		TotalPnL:      decimal.Zero,
	}
	return s.repo.Create(account)
}

// GetAccount retrieves an account by user ID.
func (s *service) GetAccount(userID uint) (*Account, error) {
	return s.repo.GetByUserID(userID)
}

// Deposit adds funds to a user's account. Amount must be positive.
func (s *service) Deposit(userID uint, amount decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return errors.New(errors.CodeInvalidInput, "amount must be positive")
	}
	account, err := s.repo.GetByUserID(userID)
	if err != nil {
		return errors.New(errors.CodeNotFound, "account not found")
	}
	account.Balance = account.Balance.Add(amount)
	account.TotalDeposit = account.TotalDeposit.Add(amount)
	return s.repo.Update(account)
}

// Withdraw removes funds from a user's account. Amount must be positive and
// cannot exceed available balance.
func (s *service) Withdraw(userID uint, amount decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return errors.New(errors.CodeInvalidInput, "amount must be positive")
	}
	account, err := s.repo.GetByUserID(userID)
	if err != nil {
		return errors.New(errors.CodeNotFound, "account not found")
	}
	if account.Balance.Cmp(amount) < 0 {
		return errors.New(errors.CodeInsufficientFunds, "insufficient balance")
	}
	account.Balance = account.Balance.Sub(amount)
	account.TotalWithdraw = account.TotalWithdraw.Add(amount)
	return s.repo.Update(account)
}

// Lock reserves funds for an active trade. Amount must be positive and
// cannot exceed available balance. Locked funds are deducted from the
// available balance and returned via Unlock after settlement.
func (s *service) Lock(userID uint, amount decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return errors.New(errors.CodeInvalidInput, "amount must be positive")
	}
	account, err := s.repo.GetByUserID(userID)
	if err != nil {
		return errors.New(errors.CodeNotFound, "account not found")
	}
	if account.Balance.Cmp(amount) < 0 {
		return errors.New(errors.CodeInsufficientFunds, "insufficient balance")
	}
	account.Balance = account.Balance.Sub(amount)
	return s.repo.Update(account)
}

// Unlock returns previously locked funds back to available balance.
// Amount must be positive.
func (s *service) Unlock(userID uint, amount decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return errors.New(errors.CodeInvalidInput, "amount must be positive")
	}
	account, err := s.repo.GetByUserID(userID)
	if err != nil {
		return errors.New(errors.CodeNotFound, "account not found")
	}
	account.Balance = account.Balance.Add(amount)
	return s.repo.Update(account)
}

// AddProfit adds profit from a settled trade to the user's balance.
// Amount must be positive.
func (s *service) AddProfit(userID uint, amount decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return errors.New(errors.CodeInvalidInput, "amount must be positive")
	}
	account, err := s.repo.GetByUserID(userID)
	if err != nil {
		return errors.New(errors.CodeNotFound, "account not found")
	}
	account.Balance = account.Balance.Add(amount)
	account.TotalPnL = account.TotalPnL.Add(amount)
	return s.repo.Update(account)
}
