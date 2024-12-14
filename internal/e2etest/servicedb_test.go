package service_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/MikeRez0/ypgophermart/internal/adapter/auth"
	"github.com/MikeRez0/ypgophermart/internal/adapter/config"
	"github.com/MikeRez0/ypgophermart/internal/adapter/logger"
	"github.com/MikeRez0/ypgophermart/internal/adapter/storage"
	"github.com/MikeRez0/ypgophermart/internal/adapter/storage/repository"
	"github.com/MikeRez0/ypgophermart/internal/core/domain"
	"github.com/MikeRez0/ypgophermart/internal/core/port"
	"github.com/MikeRez0/ypgophermart/internal/core/port/mock"
	"github.com/MikeRez0/ypgophermart/internal/core/service"
	"github.com/MikeRez0/ypgophermart/internal/e2etest/testdb"
	"github.com/golang/mock/gomock"
	"github.com/govalues/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var dbtest *testdb.TestDBInstance

func setup() {
	var err error
	dbtest, err = testdb.NewTestDBInstance()
	if err != nil {
		log.Fatal(err)
	}
}
func shutdown() {
	if dbtest != nil {
		dbtest.Down()
	}
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func getDeps() (port.Repository, port.TokenService, error) {
	db, err := storage.NewDBStorage(context.Background(), &config.Database{DSN: dbtest.DSN})
	if err != nil {
		return nil, nil, err
	}
	err = db.RunMigrations()
	if err != nil {
		return nil, nil, err
	}
	repo, err := repository.NewRepository(db)
	if err != nil {
		return nil, nil, err
	}
	ts, err := auth.New()
	if err != nil {
		return nil, nil, err
	}

	return repo, ts, nil
}

func TestServiceDB_UserRegister(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger, _ := zap.NewProduction()

	type userRegisterTest struct {
		name      string
		user      domain.User
		expError  error
		expResult *domain.User
	}

	tests := []userRegisterTest{
		{
			name:      "Register good",
			user:      domain.User{Login: "test", Password: "test"},
			expError:  nil,
			expResult: &domain.User{Login: "test"},
		},
		{
			name:      "Register good",
			user:      domain.User{Login: "test2", Password: "test"},
			expError:  nil,
			expResult: &domain.User{Login: "test2"},
		},
		{
			name:      "Register already exists",
			user:      domain.User{Login: "test", Password: "test"},
			expError:  domain.ErrConflictingData,
			expResult: nil,
		},
	}

	repo, ts, err := getDeps()
	assert.NoError(t, err)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			accrual := mock.NewMockAccrualServiceClient(mockCtrl)

			s, err := service.NewService(repo, ts, accrual, logger)
			assert.NoError(t, err)

			result, err := s.RegisterUser(context.Background(), &test.user)

			if test.expResult != nil {
				assert.Equal(t, test.expResult.Login, result.Login)
			}
			assert.Equal(t, test.expError, err)
		})
	}
}

func TestServiceDB_UserLogin(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger, _ := zap.NewProduction()

	type userLoginTest struct {
		name         string
		registerUser bool
		user         domain.User
		expError     error
	}

	tests := []userLoginTest{
		{
			registerUser: true,
			name:         "Login good",
			user:         domain.User{Login: "test", Password: "test"},

			expError: nil,
		},
		{
			name:     "Password bad",
			user:     domain.User{Login: "test", Password: "hacker"},
			expError: domain.ErrInvalidCredentials,
		},
		{
			name:     "Login bad",
			user:     domain.User{Login: "hacker", Password: "test"},
			expError: domain.ErrInvalidCredentials,
		},
	}

	repo, ts, err := getDeps()
	assert.NoError(t, err)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			accrual := mock.NewMockAccrualServiceClient(mockCtrl)

			s, err := service.NewService(repo, ts, accrual, logger)
			assert.NoError(t, err)

			if test.registerUser {
				s.RegisterUser(context.Background(), &test.user)
			}

			token, err := s.LoginUser(context.Background(), test.user.Login, test.user.Password)
			assert.Equal(t, test.expError, err)

			if token != "" {
				_, err := ts.VerifyToken(token)
				assert.NoError(t, err)
			}
		})
	}
}

func TestServiceDB_CreateOrder(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger, _ := zap.NewProduction()

	type createOrderTest struct {
		name      string
		order     domain.Order
		mock      func(as *mock.MockAccrualServiceClient)
		expError  error
		expResult domain.Order
	}

	tests := []createOrderTest{
		{
			name:  "Create good order",
			order: domain.Order{Number: "125", UserID: 1},
			mock: func(accrual *mock.MockAccrualServiceClient) {
				accrual.EXPECT().ScheduleOrderAccrual(domain.OrderNumber("125"))
			},
			expError: nil,
		},
		{
			name:  "Create good order 2",
			order: domain.Order{Number: "12345678903", UserID: 2},
			mock: func(accrual *mock.MockAccrualServiceClient) {
				accrual.EXPECT().ScheduleOrderAccrual(domain.OrderNumber("12345678903"))
			},
			expError: nil,
		},
		{
			name:     "Order already exist",
			order:    domain.Order{Number: "125", UserID: 1},
			expError: domain.ErrOrderAlreadyAcceptedByUser,
		},
		{
			name:     "Order already exist (other user)",
			order:    domain.Order{Number: "125", UserID: 2},
			expError: domain.ErrOrderAlreadyAcceptedBAnotherUser,
		},
		{
			name:     "Order bad number",
			order:    domain.Order{Number: "123", UserID: 1},
			expError: domain.ErrOrderBadNumber,
		},
	}

	repo, ts, err := getDeps()
	assert.NoError(t, err)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			accrual := mock.NewMockAccrualServiceClient(mockCtrl)
			if test.mock != nil {
				test.mock(accrual)
			}

			s, err := service.NewService(repo, ts, accrual, logger)
			assert.NoError(t, err)

			result, err := s.CreateOrder(context.Background(), &test.order)

			if result != nil {
				assert.Equal(t, test.order.Number, result.Number)
				assert.Equal(t, test.order.UserID, result.UserID)
			}
			assert.Equal(t, test.expError, err)
		})
	}
}

func TestServiceDb_AccrualWithdrawal(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	l := logger.NewLogger(&config.App{LogLevel: "debug"})
	assert.NotNil(t, l)

	type accrualWithdrawalTest struct {
		name        string
		orderNumber domain.OrderNumber
		amount      decimal.Decimal
		accrual     bool
		expError    error
		expResult   domain.Balance
	}

	tests := []accrualWithdrawalTest{
		{
			name:        "accrual success",
			orderNumber: domain.OrderNumber("2377225624"),
			amount:      decimal.Hundred,
			accrual:     true,
			expError:    nil,
			expResult:   domain.Balance{Current: decimal.Hundred, Withdrawn: decimal.Zero},
		},
		{
			name:        "withdrawal fail",
			orderNumber: domain.OrderNumber("1230"),
			amount:      decimal.MustParse("200"),
			accrual:     false,
			expError:    domain.ErrInsufficientBalance,
			expResult:   domain.Balance{Current: decimal.Hundred, Withdrawn: decimal.Zero},
		},
		{
			name:        "withdrawal success",
			orderNumber: domain.OrderNumber("1230"),
			amount:      decimal.Hundred,
			accrual:     false,
			expError:    nil,
			expResult:   domain.Balance{Withdrawn: decimal.Hundred},
		},
	}

	repo, ts, err := getDeps()
	assert.NoError(t, err)
	accrual := mock.NewMockAccrualServiceClient(mockCtrl)
	accrual.EXPECT().ScheduleOrderAccrual(gomock.Any())

	s, err := service.NewService(repo, ts, accrual, l)
	assert.NoError(t, err)

	user, err := s.RegisterUser(context.Background(), &domain.User{Login: "balance", Password: "test"})
	assert.NoError(t, err)

	l.Debug("Created user", zap.Any("user", user))

	ctx := context.Background()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			l.Debug(test.name+" start", zap.Time("now", time.Now()))
			b, err := s.GetUserBalance(ctx, user.ID)
			assert.NoError(t, err)
			l.Debug(test.name+" balance >", zap.Any("balance", b))

			var result *domain.Balance
			if test.accrual {
				_, err = s.CreateOrder(ctx, &domain.Order{Number: test.orderNumber, UserID: user.ID})
				result, err = s.Accrual(ctx, user.ID, test.orderNumber, test.amount)
			} else {
				result, err = s.Withdrawal(ctx, user.ID, test.orderNumber, test.amount)
			}

			if result != nil {
				assert.True(t,
					test.expResult.Withdrawn.Equal(result.Withdrawn) &&
						test.expResult.Current.Equal(result.Current),

					fmt.Sprintf("expected: %v", test.expResult),
					fmt.Sprintf("actual: %v", result))
			}
			assert.Equal(t, test.expError, err)
			l.Debug(test.name+" balance <", zap.Any("balance", result))
			l.Debug(test.name+" finished", zap.Time("now", time.Now()))
		})
	}
}
