package port

import (
	"context"

	"github.com/MikeRez0/ypgophermart/internal/core/domain"
	"github.com/govalues/decimal"
)

type Service interface {
	RegisterUser(ctx context.Context, user *domain.User) (*domain.User, error)
	LoginUser(ctx context.Context, login string, password string) (string, error)

	CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error)
	GetOrdersByUser(context.Context, uint64) ([]*domain.Order, error)

	GetUserBalance(ctx context.Context, user *domain.User) (*domain.Balance, error)
	Accrual(ctx context.Context, user *domain.User, amount decimal.Decimal) (*domain.Balance, error)
	Withdrawal(ctx context.Context, user *domain.User, amount decimal.Decimal) (*domain.Balance, error)
}
