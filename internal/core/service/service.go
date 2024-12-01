package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/MikeRez0/ypgophermart/internal/core/domain"
	"github.com/MikeRez0/ypgophermart/internal/core/port"
	"github.com/MikeRez0/ypgophermart/internal/core/utils"
	"github.com/govalues/decimal"
	"go.uber.org/zap"
)

type Service struct {
	repo         port.Repository
	tokenService port.TokenService
	accrual      port.AccrualServiceClient
	logger       *zap.Logger
}

func NewService(repo port.Repository, tokenService port.TokenService,
	accrualService port.AccrualServiceClient, logger *zap.Logger) (*Service, error) {
	return &Service{
		repo:         repo,
		tokenService: tokenService,
		accrual:      accrualService,
		logger:       logger,
	}, nil
}

func (s *Service) RegisterUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	exUser, err := s.repo.GetUserByLogin(ctx, user.Login)
	if err != nil && !errors.Is(err, domain.ErrDataNotFound) {
		s.logger.Error("Get user", zap.Error(err))
		return nil, domain.ErrInternal
	}

	if exUser != nil {
		return nil, domain.ErrConflictingData
	}

	newUser, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		s.logger.Error("Create user", zap.Error(err))
		return nil, domain.ErrInternal
	}

	return newUser, nil
}

func (s *Service) LoginUser(ctx context.Context, login string, password string) (string, error) {
	user, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, domain.ErrDataNotFound) {
			return "", domain.ErrInvalidCredentials
		}
		return "", domain.ErrInternal
	}

	err = utils.ComparePassword(password, user.Password)
	if err != nil {
		return "", domain.ErrInvalidCredentials
	}

	token, err := s.tokenService.CreateToken(user)
	if err != nil {
		s.logger.Error("Create token", zap.Error(err))
		return "", domain.ErrTokenCreation
	}

	return token, nil
}

func (s *Service) CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	// check number format by Luna
	err := utils.ValidateLuhn(strconv.Itoa(int(order.Number)))
	if err != nil {
		return nil, domain.ErrOrderBadNumber
	}

	// check existance
	exOrder, err := s.repo.ReadOrder(ctx, order.Number)
	if err != nil && (!errors.Is(err, domain.ErrDataNotFound)) {
		return nil, err
	}
	if exOrder != nil {
		if exOrder.UserID == order.UserID {
			return nil, domain.ErrOrderAlreadyAcceptedByUser
		} else {
			return nil, domain.ErrOrderAlreadyAcceptedBAnotherUser
		}
	}

	// save
	order.UploadedAt = time.Now()
	order.Status = domain.OrderStatusNew
	order.Accrual = decimal.Zero

	newOrder, err := s.repo.CreateOrder(ctx, order)
	if err != nil {
		if errors.Is(err, domain.ErrConflictingData) {
			return nil, domain.ErrOrderAlreadyAcceptedBAnotherUser
		}
		s.logger.Error("Create order", zap.Error(err))
		return nil, err
	}

	//TODO: shedule accrual
	go s.processOrder(context.Background(), newOrder)

	return newOrder, nil
}

func (s *Service) processOrder(ctx context.Context, order *domain.Order) {
	orderAccrual, err := s.accrual.GetOrderAccrual(order.Number)
	if err != nil {
		s.logger.Error("accrual error", zap.Error(err))
		return
	}
	if orderAccrual.Status == "PROCESSED" {
		order.Accrual = orderAccrual.Accrual
		order.Status = domain.OrderStatusProcessed
	} else if order.Status == "INVALID" {
		order.Status = domain.OrderStatusInvalid
	} else {
		order.Status = domain.OrderStatusProcessing
	}

	order, err = s.repo.UpdateOrder(ctx, order)
	if err != nil {
		s.logger.Error("update order error", zap.Error(err))
		return
	}

}

func (s *Service) GetOrdersByUser(ctx context.Context, userID uint64) ([]*domain.Order, error) {
	list, err := s.repo.ListOrdersByUser(ctx, userID)
	if err != nil {
		s.logger.Error("Get orders for user", zap.Error(err))
		return nil, err
	}
	return list, nil
}

func (s *Service) GetUserBalance(ctx context.Context, userID uint64) (*domain.Balance, error) {
	balance, err := s.repo.ReadBalanceByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Get balance error", zap.Error(err))
	}

	return balance, nil
}

func (s *Service) Accrual(ctx context.Context,
	userID uint64,
	orderID uint64,
	amount decimal.Decimal,
) (*domain.Balance, error) {
	order, err := s.repo.ReadOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	balance, err := s.repo.UpdateUserBalanceByOrder(ctx, userID, order.Number,
		func(b *domain.Balance, o *domain.Order) error {
			newAmount, err := b.Current.Add(amount)
			if err != nil {
				return fmt.Errorf("Math error:%w", err)
			}
			b.Current = newAmount
			order.Accrual = amount
			order.Status = domain.OrderStatusProcessed

			return nil
		})
	if err != nil {
		return nil, err
	}

	return balance, nil
}
func (s *Service) Withdrawal(ctx context.Context,
	userID uint64,
	orderID uint64,
	amount decimal.Decimal,
) (*domain.Balance, error) {
	order, err := s.repo.ReadOrder(ctx, orderID)
	if err != nil {
		if errors.Is(err, domain.ErrDataNotFound) {
			return nil, domain.ErrOrderBadNumber
		}
		return nil, err
	}

	balance, err := s.repo.UpdateUserBalanceByOrder(ctx, userID, order.Number,
		func(b *domain.Balance, o *domain.Order) error {
			if b.Current.Cmp(amount) < 0 {
				return domain.ErrInsufficientBalance
			}

			if o.Withdrawal.Cmp(decimal.Zero) != 0 {
				return domain.ErrOrderDoubleWithdraw
			}
			o.Withdrawal = amount

			newCurrent, err := b.Current.Sub(amount)
			if err != nil {
				return fmt.Errorf("math error:%w", err)
			}
			b.Current = newCurrent

			newWithdrawn, err := b.Withdrawn.Add(amount)
			if err != nil {
				return fmt.Errorf("math error:%w", err)
			}
			b.Withdrawn = newWithdrawn

			return nil
		})
	if err != nil {
		return nil, err
	}

	return balance, nil
}
