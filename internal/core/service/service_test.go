package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/MikeRez0/ypgophermart/internal/adapter/auth"
	"github.com/MikeRez0/ypgophermart/internal/core/domain"
	"github.com/MikeRez0/ypgophermart/internal/core/port/mock"
	"github.com/MikeRez0/ypgophermart/internal/core/service"
	"github.com/MikeRez0/ypgophermart/internal/core/utils"
	"github.com/golang/mock/gomock"
	"github.com/govalues/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type prepareMocks func(repo *mock.MockRepository, accrual *mock.MockAccrualServiceClient)

func TestService_UserRegister(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger, _ := zap.NewProduction()

	type userRegisterTest struct {
		name      string
		user      domain.User
		mock      prepareMocks
		expError  error
		expResult *domain.User
	}

	hashedPass, _ := utils.HashPassword("test")
	user := domain.User{
		Login:    "test",
		Password: hashedPass,
		ID:       1,
	}

	tests := []userRegisterTest{
		{
			name: "Register good",
			user: domain.User{Login: user.Login, Password: "test"},
			mock: func(repo *mock.MockRepository, accrual *mock.MockAccrualServiceClient) {
				repo.EXPECT().GetUserByLogin(gomock.Any(), user.Login).Return(nil, domain.ErrDataNotFound)
				repo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(&user, nil)
			},
			expError:  nil,
			expResult: &user,
		},
		{
			name: "Register already exists",
			user: domain.User{Login: user.Login, Password: "test"},
			mock: func(repo *mock.MockRepository, accrual *mock.MockAccrualServiceClient) {
				repo.EXPECT().GetUserByLogin(gomock.Any(), user.Login).Return(&user, nil)
			},
			expError:  domain.ErrConflictingData,
			expResult: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := mock.NewMockRepository(mockCtrl)
			ts := mock.NewMockTokenService(mockCtrl)
			accrual := mock.NewMockAccrualServiceClient(mockCtrl)
			test.mock(repo, accrual)

			s, err := service.NewService(repo, ts, accrual, logger)
			assert.NoError(t, err)

			result, err := s.RegisterUser(context.Background(), &test.user)

			assert.Equal(t, test.expResult, result)
			assert.Equal(t, test.expError, err)
		})
	}
}

func TestService_UserLogin(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger, _ := zap.NewProduction()

	type userLoginTest struct {
		name     string
		user     domain.User
		mock     prepareMocks
		expError error
	}

	hashedPass, _ := utils.HashPassword("test")
	user := domain.User{
		Login:    "test",
		Password: hashedPass,
		ID:       1,
	}

	tests := []userLoginTest{
		{
			name: "Login good",
			user: domain.User{Login: user.Login, Password: "test", ID: 1},
			mock: func(repo *mock.MockRepository, accrual *mock.MockAccrualServiceClient) {
				repo.EXPECT().GetUserByLogin(gomock.Any(), user.Login).Return(&user, nil)
			},
			expError: nil,
		},
		{
			name: "Password bad",
			user: domain.User{Login: user.Login, Password: "hacker"},
			mock: func(repo *mock.MockRepository, accrual *mock.MockAccrualServiceClient) {
				repo.EXPECT().GetUserByLogin(gomock.Any(), user.Login).Return(&user, nil)
			},
			expError: domain.ErrInvalidCredentials,
		},
		{
			name: "Login bad",
			user: domain.User{Login: "hacker", Password: "test"},
			mock: func(repo *mock.MockRepository, accrual *mock.MockAccrualServiceClient) {
				repo.EXPECT().GetUserByLogin(gomock.Any(), "hacker").Return(nil, domain.ErrDataNotFound)
			},
			expError: domain.ErrInvalidCredentials,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := mock.NewMockRepository(mockCtrl)
			ts, err := auth.New()
			assert.NoError(t, err)

			accrual := mock.NewMockAccrualServiceClient(mockCtrl)
			test.mock(repo, accrual)

			s, err := service.NewService(repo, ts, accrual, logger)
			assert.NoError(t, err)

			token, err := s.LoginUser(context.Background(), test.user.Login, test.user.Password)
			assert.Equal(t, test.expError, err)

			if token != "" {
				payload, err := ts.VerifyToken(token)
				assert.NoError(t, err)
				assert.Equal(t, payload.UserID, test.user.ID)
			}
		})
	}
}

func TestService_CreateOrder(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger, _ := zap.NewProduction()

	type createOrderTest struct {
		name      string
		order     domain.Order
		mock      prepareMocks
		expError  error
		expResult *domain.Order
	}

	order := domain.Order{
		Number:     "125",
		UserID:     1,
		UploadedAt: time.Now(),
		Status:     domain.OrderStatusNew,
		Accrual:    decimal.Zero,
		Withdrawal: decimal.Zero,
	}

	tests := []createOrderTest{
		{
			name:  "Create good order",
			order: domain.Order{Number: "125", UserID: 1},
			mock: func(repo *mock.MockRepository, accrual *mock.MockAccrualServiceClient) {
				repo.EXPECT().ReadOrder(gomock.Any(), domain.OrderNumber("125")).
					Return(nil, domain.ErrDataNotFound)
				repo.EXPECT().CreateOrder(gomock.Any(), gomock.Any()).
					Return(&order, nil)

				accrual.EXPECT().ScheduleOrderAccrual(domain.OrderNumber("125"))
			},
			expError:  nil,
			expResult: &order,
		},
		{
			name:  "Order already exist",
			order: domain.Order{Number: "125", UserID: 1},
			mock: func(repo *mock.MockRepository, accrual *mock.MockAccrualServiceClient) {
				repo.EXPECT().ReadOrder(gomock.Any(), domain.OrderNumber("125")).
					Return(&order, nil)
			},
			expError:  domain.ErrOrderAlreadyAcceptedByUser,
			expResult: nil,
		},
		{
			name:  "Order already exist",
			order: domain.Order{Number: "125", UserID: 1},
			mock: func(repo *mock.MockRepository, accrual *mock.MockAccrualServiceClient) {
				repo.EXPECT().ReadOrder(gomock.Any(), domain.OrderNumber("125")).
					Return(&domain.Order{
						Number:     "125",
						UserID:     2,
						UploadedAt: time.Now(),
						Status:     domain.OrderStatusNew,
						Accrual:    decimal.Zero,
						Withdrawal: decimal.Zero,
					}, nil)
			},
			expError:  domain.ErrOrderAlreadyAcceptedBAnotherUser,
			expResult: nil,
		},
		{
			name:  "Order bad number",
			order: domain.Order{Number: "123", UserID: 1},
			mock: func(repo *mock.MockRepository, accrual *mock.MockAccrualServiceClient) {
			},
			expError:  domain.ErrOrderBadNumber,
			expResult: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := mock.NewMockRepository(mockCtrl)
			ts := mock.NewMockTokenService(mockCtrl)
			accrual := mock.NewMockAccrualServiceClient(mockCtrl)
			test.mock(repo, accrual)

			s, err := service.NewService(repo, ts, accrual, logger)
			assert.NoError(t, err)

			result, err := s.CreateOrder(context.Background(), &test.order)

			assert.Equal(t, test.expResult, result)
			assert.Equal(t, test.expError, err)
		})
	}
}

func TestService_AccrualWithdrawal(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger, _ := zap.NewProduction()

	type accrualWithdrawalTest struct {
		name      string
		order     domain.Order
		balance   domain.Balance
		amount    decimal.Decimal
		accrual   bool
		mock      prepareMocks
		expError  error
		expResult *domain.Balance
	}

	order := domain.Order{
		Number:     "125",
		UserID:     1,
		UploadedAt: time.Now(),
		Status:     domain.OrderStatusNew,
		Accrual:    decimal.Zero,
		Withdrawal: decimal.Zero,
	}
	balance := domain.Balance{
		UserID:    1,
		Current:   decimal.MustParse("100"),
		Withdrawn: decimal.MustParse("200"),
	}

	tests := []accrualWithdrawalTest{
		{
			name:    "accrual success",
			order:   order,
			balance: balance,
			amount:  decimal.Hundred,
			accrual: true,
			mock: func(repo *mock.MockRepository, accrual *mock.MockAccrualServiceClient) {
				repo.EXPECT().ReadOrder(gomock.Any(), order.Number).
					Return(&order, nil)
				repo.EXPECT().UpdateUserBalanceByOrder(context.Background(),
					gomock.Any(), false,
					gomock.Any(),
				).Return(&balance, nil)
			},
			expError:  nil,
			expResult: &balance,
		},
		{
			name:    "withdrawal success",
			order:   order,
			balance: balance,
			amount:  decimal.Hundred,
			accrual: false,
			mock: func(repo *mock.MockRepository, accrual *mock.MockAccrualServiceClient) {
				repo.EXPECT().ReadOrder(gomock.Any(), order.Number).
					Return(nil, domain.ErrDataNotFound)
				repo.EXPECT().UpdateUserBalanceByOrder(context.Background(),
					gomock.Any(), true,
					gomock.Any(),
				).Return(&balance, nil)
			},
			expError:  nil,
			expResult: &balance,
		},
		{
			name:    "withdrawal fail",
			order:   order,
			balance: balance,
			amount:  decimal.MustParse("200"),
			accrual: false,
			mock: func(repo *mock.MockRepository, accrual *mock.MockAccrualServiceClient) {
				repo.EXPECT().ReadOrder(gomock.Any(), order.Number).
					Return(nil, domain.ErrDataNotFound)
				repo.EXPECT().UpdateUserBalanceByOrder(context.Background(),
					gomock.Any(), true,
					gomock.Any(),
				).Return(nil, domain.ErrInsufficientBalance)
			},
			expError:  domain.ErrInsufficientBalance,
			expResult: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := mock.NewMockRepository(mockCtrl)
			ts := mock.NewMockTokenService(mockCtrl)
			accrual := mock.NewMockAccrualServiceClient(mockCtrl)
			test.mock(repo, accrual)

			s, err := service.NewService(repo, ts, accrual, logger)
			assert.NoError(t, err)

			var result *domain.Balance
			if test.accrual {
				result, err = s.Accrual(context.Background(), test.balance.UserID, test.order.Number, test.amount)
			} else {
				result, err = s.Withdrawal(context.Background(), test.balance.UserID, test.order.Number, test.amount)
			}

			if test.expResult == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, *test.expResult, *result)
			}
			assert.Equal(t, test.expError, err)
		})
	}
}
