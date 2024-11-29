package service

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/MikeRez0/ypgophermart/internal/core/domain"
	"github.com/MikeRez0/ypgophermart/internal/core/port"
	"github.com/MikeRez0/ypgophermart/internal/core/utils"
	"github.com/govalues/decimal"
)

type Service struct {
	repo         port.Repository
	tokenService port.TokenService
}

func NewService(repo port.Repository, tokenService port.TokenService) (*Service, error) {
	return &Service{repo: repo, tokenService: tokenService}, nil
}

func (us *Service) RegisterUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	exUser, err := us.repo.GetUserByLogin(ctx, user.Login)
	if err != nil && !errors.Is(err, domain.ErrDataNotFound) {
		return nil, domain.ErrInternal
	}

	if exUser != nil {
		return nil, domain.ErrConflictingData
	}

	newUser, err := us.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, domain.ErrInternal
	}

	return newUser, nil
}

func (us *Service) LoginUser(ctx context.Context, login string, password string) (string, error) {
	user, err := us.repo.GetUserByLogin(ctx, login)
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

	token, err := us.tokenService.CreateToken(user)
	if err != nil {
		return "", domain.ErrTokenCreation
	}

	return token, nil
}

func (os *Service) CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	// check number format by Luna
	err := utils.ValidateLuhn(strconv.Itoa(int(order.Number)))
	if err != nil {
		return nil, domain.ErrOrderBadNumber
	}

	// check existance
	exOrder, err := os.repo.ReadOrder(ctx, order.Number)
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

	newOrder, err := os.repo.CreateOrder(ctx, order)
	if err != nil {
		if errors.Is(err, domain.ErrConflictingData) {
			return nil, domain.ErrOrderAlreadyAcceptedBAnotherUser
		}

		return nil, err
	}

	//TODO: shedule accrual

	return newOrder, nil
}

func (s *Service) GetOrdersByUser(ctx context.Context, userID uint64) ([]*domain.Order, error) {
	list, err := s.repo.ListOrdersByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (us *Service) GetUserBalance(ctx context.Context, user *domain.User) (*domain.Balance, error) {
	return nil, nil
}

func (us *Service) Accrual(ctx context.Context, user *domain.User, amount decimal.Decimal) (*domain.Balance, error) {
	return nil, nil
}
func (us *Service) Withdrawal(ctx context.Context, user *domain.User, amount decimal.Decimal) (*domain.Balance, error) {
	return nil, nil
}
