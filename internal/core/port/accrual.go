package port

import (
	"context"

	"github.com/MikeRez0/ypgophermart/internal/core/domain"
	"github.com/govalues/decimal"
)

//go:generate mockgen -source=accrual.go -destination=mock/accrual.go -package=mock
type AccrualServiceClient interface {
	ScheduleOrderAccrual(orderNumber domain.OrderNumber)
}

type OrderAccrualUpdater interface {
	AccrualOrder(ctx context.Context, orderNumber domain.OrderNumber, amount decimal.Decimal) error
	UpdateOrderStatus(ctx context.Context, orderNumber domain.OrderNumber, status domain.OrderStatus) error
}
