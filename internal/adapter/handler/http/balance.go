package http

import (
	"time"

	"github.com/MikeRez0/ypgophermart/internal/core/domain"
	"github.com/MikeRez0/ypgophermart/internal/core/port"
	"github.com/gin-gonic/gin"
	"github.com/govalues/decimal"
	"go.uber.org/zap"
)

type BalanceHandler struct {
	Handler
	service port.Service
}

func NewBalanceHandler(service port.Service, logger *zap.Logger) (*BalanceHandler, error) {
	return &BalanceHandler{
		Handler: Handler{logger: logger},
		service: service,
	}, nil
}

type balanceResponse struct {
	Current   jsonDecimal `json:"current"`
	Withdrawn jsonDecimal `json:"withdrawn"`
}

func (bh *BalanceHandler) UserBalance(ctx *gin.Context) {
	userID := getAuthPayload(ctx).UserID

	balance, err := bh.service.GetUserBalance(ctx, userID)
	if err != nil {
		bh.handleError(ctx, err)
		return
	}

	bh.handleSuccess(ctx, balanceResponse{
		Current:   jsonDecimal(balance.Current),
		Withdrawn: jsonDecimal(balance.Withdrawn)})
}

type withdrawnRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func (bh *BalanceHandler) Withdraw(ctx *gin.Context) {
	req := withdrawnRequest{}
	err := ctx.ShouldBindBodyWithJSON(&req)
	if err != nil {
		bh.handleValidationError(ctx, err)
		return
	}

	userID := getAuthPayload(ctx).UserID

	amount, err := decimal.NewFromFloat64(req.Sum)
	if err != nil {
		bh.handleValidationError(ctx, err)
		return
	}

	balance, err := bh.service.Withdrawal(ctx, userID, domain.OrderNumber(req.Order), amount)
	if err != nil {
		bh.handleError(ctx, err)
		return
	}
	bh.handleSuccess(ctx,
		balanceResponse{
			Current:   jsonDecimal(balance.Current),
			Withdrawn: jsonDecimal(balance.Withdrawn),
		})
}

type withdrawalResponse struct {
	Order       string      `json:"order"`
	Sum         jsonDecimal `json:"sum"`
	ProcessedAt time.Time   `json:"processed_at"`
}

func (bh *BalanceHandler) ListWithdrawals(ctx *gin.Context) {
	userID := getAuthPayload(ctx).UserID

	list, err := bh.service.GetOrdersByUser(ctx, userID)
	if err != nil {
		bh.handleError(ctx, err)
		return
	}

	result := make([]withdrawalResponse, 0, len(list))
	for _, i := range list {
		if i.Withdrawal.Cmp(decimal.Zero) == 0 {
			continue
		}
		result = append(result, withdrawalResponse{
			Order:       string(i.Number),
			Sum:         jsonDecimal(i.Withdrawal),
			ProcessedAt: i.UploadedAt,
		})
	}

	bh.handleSuccess(ctx, result)
}
