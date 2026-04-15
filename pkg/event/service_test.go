package event

import (
	"testing"
	"time"

	"github.com/huinong/golang-claude/pkg/common/decimal"
	"github.com/huinong/golang-claude/pkg/common/errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestEventService(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	db.AutoMigrate(&Event{})

	repo := NewGormRepository(db)
	service := NewService(repo)

	tests := []struct {
		name    string
		action  func() (*Event, error)
		wantErr bool
		errCode int
		check   func(*testing.T, *Event)
	}{
		{
			name: "create event success",
			action: func() (*Event, error) {
				startTime := time.Now().Add(-1 * time.Hour)
				endTime := time.Now().Add(1 * time.Hour)
				return service.Create(
					"Will it rain tomorrow?",
					"Test description",
					startTime, endTime,
					decimal.NewFromInt(60).Div(decimal.NewFromInt(100)), // 60%
					decimal.NewFromInt(1000), // 1000 shares
				)
			},
			wantErr: false,
			check: func(t *testing.T, event *Event) {
				if event.Status != StatusPending {
					t.Errorf("expected status PENDING, got %s", event.Status)
				}
				// Check inventory calculation:
				// initialYesPrice 0.6, initialSupply 1000
				// YesInventory = 1000 * (1 - 0.6) = 400
				// NoInventory = 1000 * 0.6 = 600
				expectedYesInv := decimal.NewFromInt(400)
				if !event.YesInventory.Equal(expectedYesInv) {
					t.Errorf("expected yes inventory %v, got %v", expectedYesInv, event.YesInventory)
				}
				expectedNoInv := decimal.NewFromInt(600)
				if !event.NoInventory.Equal(expectedNoInv) {
					t.Errorf("expected no inventory %v, got %v", expectedNoInv, event.NoInventory)
				}
				expectedTotal := decimal.NewFromInt(1000)
				if !event.TotalInventory().Equal(expectedTotal) {
					t.Errorf("expected total inventory %v, got %v", expectedTotal, event.TotalInventory())
				}
				expectedNoPrice := decimal.NewFromInt(40).Div(decimal.NewFromInt(100))
				if !event.NoPrice().Equal(expectedNoPrice) {
					t.Errorf("expected no price %v, got %v", expectedNoPrice, event.NoPrice())
				}
			},
		},
		{
			name: "create event empty title should fail",
			action: func() (*Event, error) {
				startTime := time.Now()
				endTime := time.Now().Add(24 * time.Hour)
				return service.Create(
					"",
					"Description",
					startTime, endTime,
					decimal.NewFromInt(50).Div(decimal.NewFromInt(100)),
					decimal.NewFromInt(1000),
				)
			},
			wantErr: true,
			errCode: errors.CodeInvalidInput,
		},
		{
			name: "create event title too long should fail",
			action: func() (*Event, error) {
				longTitle := make([]byte, 201)
				for i := range longTitle {
					longTitle[i] = 'a'
				}
				startTime := time.Now()
				endTime := time.Now().Add(24 * time.Hour)
				return service.Create(
					string(longTitle),
					"Description",
					startTime, endTime,
					decimal.NewFromInt(50).Div(decimal.NewFromInt(100)),
					decimal.NewFromInt(1000),
				)
			},
			wantErr: true,
			errCode: errors.CodeInvalidInput,
		},
		{
			name: "create event end time before start time should fail",
			action: func() (*Event, error) {
				startTime := time.Now().Add(24 * time.Hour)
				endTime := time.Now()
				return service.Create(
					"Test Event",
					"Description",
					startTime, endTime,
					decimal.NewFromInt(50).Div(decimal.NewFromInt(100)),
					decimal.NewFromInt(1000),
				)
			},
			wantErr: true,
			errCode: errors.CodeInvalidInput,
		},
		{
			name: "create event price out of range should fail (0.001)",
			action: func() (*Event, error) {
				startTime := time.Now()
				endTime := time.Now().Add(24 * time.Hour)
				return service.Create(
					"Test Event",
					"Description",
					startTime, endTime,
					decimal.NewFromInt(1).Div(decimal.NewFromInt(1000)),
					decimal.NewFromInt(1000),
				)
			},
			wantErr: true,
			errCode: errors.CodeInvalidInput,
		},
		{
			name: "create event zero initial supply should fail",
			action: func() (*Event, error) {
				startTime := time.Now()
				endTime := time.Now().Add(24 * time.Hour)
				return service.Create(
					"Test Event",
					"Description",
					startTime, endTime,
					decimal.NewFromInt(50).Div(decimal.NewFromInt(100)),
					decimal.Zero,
				)
			},
			wantErr: true,
			errCode: errors.CodeInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := tt.action()
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				var appErr *errors.AppError
				appErr, ok := err.(*errors.AppError)
				if !ok {
					t.Errorf("expected errors.AppError, got %T", err)
				} else if appErr.Code != tt.errCode {
					t.Errorf("error code = %d, want %d", appErr.Code, tt.errCode)
				}
				return
			}

			tt.check(t, event)
		})
	}
}

func TestStartTradingAndSettle(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	db.AutoMigrate(&Event{})

	repo := NewGormRepository(db)
	service := NewService(repo)

	// Create event
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now().Add(1 * time.Hour)
	event, err := service.Create(
		"Test Event",
		"Description",
		startTime, endTime,
		decimal.NewFromInt(50).Div(decimal.NewFromInt(100)),
		decimal.NewFromInt(1000),
	)
	if err != nil {
		t.Fatalf("create event failed: %v", err)
	}

	// Start trading
	err = service.StartTrading(event.ID)
	if err != nil {
		t.Fatalf("start trading failed: %v", err)
	}

	// Check status
	event, err = service.GetByID(event.ID)
	if err != nil {
		t.Fatalf("get event failed: %v", err)
	}
	if event.Status != StatusTrading {
		t.Errorf("expected status TRADING, got %s", event.Status)
	}

	// Cannot start trading again
	err = service.StartTrading(event.ID)
	if err == nil {
		t.Errorf("expected error when starting already started event, got nil")
	}
	var appErr *errors.AppError
	appErr, ok := err.(*errors.AppError)
	if !ok || appErr.Code != errors.CodeInvalidInput {
		t.Errorf("expected CodeInvalidInput error, got %v", err)
	}

	// Settle event
	err = service.Settle(event.ID, ResultYesWon)
	if err != nil {
		t.Fatalf("settle failed: %v", err)
	}

	// Check result
	event, err = service.GetByID(event.ID)
	if err != nil {
		t.Fatalf("get event failed: %v", err)
	}
	if event.Status != StatusSettled {
		t.Errorf("expected status SETTLED, got %s", event.Status)
	}
	if event.Result != ResultYesWon {
		t.Errorf("expected result YES_WON, got %s", event.Result)
	}
}

func TestListActive(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	db.AutoMigrate(&Event{})

	repo := NewGormRepository(db)
	service := NewService(repo)

	// Create three events
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now().Add(1 * time.Hour)

	_, _ = service.Create("Event 1", "Desc", startTime, endTime, decimal.NewFromInt(50).Div(decimal.NewFromInt(100)), decimal.NewFromInt(1000))
	event2, _ := service.Create("Event 2", "Desc", startTime, endTime, decimal.NewFromInt(50).Div(decimal.NewFromInt(100)), decimal.NewFromInt(1000))
	_, _ = service.Create("Event 3", "Desc", startTime, endTime, decimal.NewFromInt(50).Div(decimal.NewFromInt(100)), decimal.NewFromInt(1000))

	// Start event 2
	_ = service.StartTrading(event2.ID)

	events, err := service.ListActive()
	if err != nil {
		t.Fatalf("list active failed: %v", err)
	}

	if len(events) != 1 {
		t.Errorf("expected 1 active event, got %d", len(events))
	}
	if len(events) > 0 && events[0].ID != event2.ID {
		t.Errorf("expected event %d to be active, got %d", event2.ID, events[0].ID)
	}
}

func TestUpdatePrice(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	db.AutoMigrate(&Event{})

	repo := NewGormRepository(db)
	service := NewService(repo)

	startTime := time.Now()
	endTime := time.Now().Add(24 * time.Hour)
	event, _ := service.Create("Test", "Desc", startTime, endTime, decimal.NewFromInt(50).Div(decimal.NewFromInt(100)), decimal.NewFromInt(1000))

	newPrice := decimal.NewFromInt(65).Div(decimal.NewFromInt(100))
	err = service.UpdatePrice(event.ID, newPrice)
	if err != nil {
		t.Fatalf("update price failed: %v", err)
	}

	updated, _ := service.GetByID(event.ID)
	if !updated.YesPrice.Equal(newPrice) {
		t.Errorf("expected price %v, got %v", newPrice, updated.YesPrice)
	}
}
