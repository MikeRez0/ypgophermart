package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/MikeRez0/ypgophermart/internal/core/domain"
	"github.com/MikeRez0/ypgophermart/internal/core/port"
	"github.com/gin-gonic/gin"
	"github.com/govalues/decimal"
)

type OrderHandler struct {
	service port.Service
}

func NewOrderHandler(service port.Service) (*OrderHandler, error) {
	return &OrderHandler{
		service: service,
	}, nil
}

func (oh *OrderHandler) CreateOrder(ctx *gin.Context) {
	userID := getAuthPayload(ctx).UserID

	orderNum, err := strconv.Atoi(ctx.Param("order"))
	if err != nil {
		handleValidationError(ctx, domain.ErrOrderBadNumber)
	}

	order := &domain.Order{Number: uint64(orderNum), UserID: userID}
	_, err = oh.service.CreateOrder(ctx, order)
	if err != nil {
		handleError(ctx, err)
		return
	}

	handleSuccessWithStatus(ctx, nil, http.StatusAccepted)
}

type OrderResp struct {
	Number     uint64          `json:"number"`
	Status     string          `json:"status"`
	Accrual    decimal.Decimal `json:"accrual"`
	UploadedAt time.Time       `json:"uploaded_at"`
}

func (oh *OrderHandler) ListOrdersByUser(ctx *gin.Context) {
	// GetUser
	userID := getAuthPayload(ctx).UserID

	list, err := oh.service.GetOrdersByUser(ctx, userID)
	if err != nil {
		handleError(ctx, err)
		return
	}

	result := make([]OrderResp, 0, len(list))
	for _, o := range list {
		result = append(result, OrderResp{
			Number:     o.Number,
			Accrual:    o.Accrual,
			Status:     string(o.Status),
			UploadedAt: o.UploadedAt,
		})
	}

	handleSuccess(ctx, result)
}
