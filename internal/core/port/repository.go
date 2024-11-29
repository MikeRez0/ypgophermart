package port

import (
	"context"

	"github.com/MikeRez0/ypgophermart/internal/core/domain"
)

type Repository interface {
	//User
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	GetUserByLogin(ctx context.Context, login string) (*domain.User, error)

	// Order
	CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error)
	UpdateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error)
	ReadOrder(ctx context.Context, orderID uint64) (*domain.Order, error)
	ListOrdersByUser(ctx context.Context, userID uint64) ([]*domain.Order, error)
	ListOrdersByStatus(ctx context.Context, status domain.OrderStatus) ([]*domain.Order, error)

	//Balance
	ReadBalanceByUserID(ctx context.Context, userID uint64) (*domain.Balance, error)
	UpdateBalance(ctx context.Context, balance *domain.Balance) (*domain.Balance, error)
}
