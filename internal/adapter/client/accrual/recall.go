package accrual

import (
	"context"

	"github.com/MikeRez0/ypgophermart/internal/core/domain"
	"github.com/MikeRez0/ypgophermart/internal/core/port"
)

func RecallOrders(ctx context.Context, repo port.Repository, accrual port.AccrualServiceClient) error {
	orders, err := repo.ListOrdersByStatus(ctx, []domain.OrderStatus{domain.OrderStatusNew, domain.OrderStatusProcessing})
	if err != nil {
		return err
	}
	for _, order := range orders {
		accrual.ScheduleOrderAccrual(order.Number)
	}

	return nil
}
