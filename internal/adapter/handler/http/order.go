package http

import (
	"bytes"
	"net/http"
	"time"

	"github.com/MikeRez0/ypgophermart/internal/core/domain"
	"github.com/MikeRez0/ypgophermart/internal/core/port"
	"github.com/gin-gonic/gin"
	"github.com/govalues/decimal"
	"go.uber.org/zap"
)

type OrderHandler struct {
	Handler
	service port.Service
}

func NewOrderHandler(service port.Service, logger *zap.Logger) (*OrderHandler, error) {
	return &OrderHandler{
		Handler: *NewHandler(logger),
		service: service,
	}, nil
}

func (oh *OrderHandler) CreateOrder(ctx *gin.Context) {
	userID := getAuthPayload(ctx).UserID

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(ctx.Request.Body)
	if err != nil {
		oh.handleValidationError(ctx, domain.ErrBadRequest)
		return
	}
	defer ctx.Request.Body.Close()

	orderNum := domain.OrderNumber(buf.String())

	order := &domain.Order{Number: orderNum, UserID: userID}
	_, err = oh.service.CreateOrder(ctx, order)
	if err != nil {
		oh.handleError(ctx, err)
		return
	}

	oh.handleSuccessWithStatus(ctx, nil, http.StatusAccepted)
}

type OrderResp struct {
	Number     string           `json:"number"`
	Status     string           `json:"status"`
	Accrual    *decimal.Decimal `json:"accrual,omitempty"`
	UploadedAt time.Time        `json:"uploaded_at"`
}

func (oh *OrderHandler) ListOrdersByUser(ctx *gin.Context) {
	// GetUser
	userID := getAuthPayload(ctx).UserID

	list, err := oh.service.GetOrdersByUser(ctx, userID)
	if err != nil {
		oh.handleError(ctx, err)
		return
	}

	result := make([]OrderResp, 0, len(list))
	for _, o := range list {
		r := OrderResp{
			Number:     string(o.Number),
			Status:     string(o.Status),
			UploadedAt: o.UploadedAt,
		}
		if o.Accrual.Cmp(decimal.Zero) != 0 {
			d := o.Accrual
			r.Accrual = &d
		} else {
			r.Accrual = nil
		}
		result = append(result, r)
	}

	oh.handleSuccess(ctx, result)
}
