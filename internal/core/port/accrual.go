package port

import (
	"context"

	"github.com/MikeRez0/ypgophermart/internal/core/domain"
	"github.com/govalues/decimal"
)

type AccrualServiceClient interface {
	ScheduleOrderAccrual(orderNumber uint64)
}

type OrderAccrualUpdater interface {
	AccrualOrder(ctx context.Context, orderNumber uint64, amount decimal.Decimal) error
	UpdateOrderStatus(ctx context.Context, orderNumber uint64, status domain.OrderStatus) error
}
