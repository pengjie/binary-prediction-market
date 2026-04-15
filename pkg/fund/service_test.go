package fund

import (
	stderrors "errors"
	"testing"

	"github.com/huinong/golang-claude/pkg/common/decimal"
	"github.com/huinong/golang-claude/pkg/common/errors"
	"github.com/huinong/golang-claude/pkg/user"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestFundService(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	db.AutoMigrate(&user.User{}, &Account{})

	userRepo := user.NewGormRepository(db)
	userService := user.NewUserService(userRepo)
	u, _ := userService.Create("testfund")

	repo := NewGormRepository(db)
	service := NewService(repo)
	err = service.CreateAccount(u)
	if err != nil {
		t.Fatalf("create account failed: %v", err)
	}

	tests := []struct {
		name     string
		action   func() error
		wantErr  bool
		errCode  int
		check    func(*testing.T, *Account)
	}{
		{
			name: "deposit 1000 success",
			action: func() error {
				return service.Deposit(u.ID, decimal.NewFromInt(1000))
			},
			wantErr: false,
			check: func(t *testing.T, acc *Account) {
				want := decimal.NewFromInt(1000)
				if !acc.Balance.Equal(want) {
					t.Errorf("balance = %v, want %v", acc.Balance, want)
				}
				if !acc.TotalDeposit.Equal(want) {
					t.Errorf("totalDeposit = %v, want %v", acc.TotalDeposit, want)
				}
			},
		},
		{
			name: "deposit zero amount should fail",
			action: func() error {
				return service.Deposit(u.ID, decimal.Zero)
			},
			wantErr: true,
			errCode: errors.CodeInvalidInput,
		},
		{
			name: "withdraw 300 success",
			action: func() error {
				return service.Withdraw(u.ID, decimal.NewFromInt(300))
			},
			wantErr: false,
			check: func(t *testing.T, acc *Account) {
				wantBalance := decimal.NewFromInt(700)
				wantTotalWithdraw := decimal.NewFromInt(300)
				if !acc.Balance.Equal(wantBalance) {
					t.Errorf("balance = %v, want %v", acc.Balance, wantBalance)
				}
				if !acc.TotalWithdraw.Equal(wantTotalWithdraw) {
					t.Errorf("totalWithdraw = %v, want %v", acc.TotalWithdraw, wantTotalWithdraw)
				}
			},
		},
		{
			name: "withdraw more than balance should fail",
			action: func() error {
				return service.Withdraw(u.ID, decimal.NewFromInt(800))
			},
			wantErr: true,
			errCode: errors.CodeInsufficientFunds,
		},
		{
			name: "withdraw zero amount should fail",
			action: func() error {
				return service.Withdraw(u.ID, decimal.Zero)
			},
			wantErr: true,
			errCode: errors.CodeInvalidInput,
		},
		{
			name: "lock 500 success",
			action: func() error {
				return service.Lock(u.ID, decimal.NewFromInt(500))
			},
			wantErr: false,
			check: func(t *testing.T, acc *Account) {
				want := decimal.NewFromInt(200)
				if !acc.Balance.Equal(want) {
					t.Errorf("balance = %v, want %v", acc.Balance, want)
				}
			},
		},
		{
			name: "lock more than available should fail",
			action: func() error {
				return service.Lock(u.ID, decimal.NewFromInt(300))
			},
			wantErr: true,
			errCode: errors.CodeInsufficientFunds,
		},
		{
			name: "unlock 500 success",
			action: func() error {
				return service.Unlock(u.ID, decimal.NewFromInt(500))
			},
			wantErr: false,
			check: func(t *testing.T, acc *Account) {
				want := decimal.NewFromInt(700)
				if !acc.Balance.Equal(want) {
					t.Errorf("balance = %v, want %v", acc.Balance, want)
				}
			},
		},
		{
			name: "add profit 200 success",
			action: func() error {
				return service.AddProfit(u.ID, decimal.NewFromInt(200))
			},
			wantErr: false,
			check: func(t *testing.T, acc *Account) {
				wantBalance := decimal.NewFromInt(900)
				wantPnL := decimal.NewFromInt(200)
				if !acc.Balance.Equal(wantBalance) {
					t.Errorf("balance = %v, want %v", acc.Balance, wantBalance)
				}
				if !acc.TotalPnL.Equal(wantPnL) {
					t.Errorf("totalPnL = %v, want %v", acc.TotalPnL, wantPnL)
				}
			},
		},
		{
			name: "operations on non-existent user should fail",
			action: func() error {
				return service.Deposit(9999, decimal.NewFromInt(100))
			},
			wantErr: true,
			errCode: errors.CodeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.action()
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				var appErr *errors.AppError
				if ok := stderrors.As(err, &appErr); !ok {
					t.Errorf("expected AppError, got %T", err)
				} else if appErr.Code != tt.errCode {
					t.Errorf("error code = %d, want %d", appErr.Code, tt.errCode)
				}
				return
			}

			acc, err := service.GetAccount(u.ID)
			if err != nil {
				t.Fatalf("get account failed: %v", err)
			}
			tt.check(t, acc)
		})
	}
}

