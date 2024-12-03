package http

import (
	"fmt"
	"net/http"

	"github.com/MikeRez0/ypgophermart/internal/core/domain"
	"github.com/gin-gonic/gin"
	"github.com/govalues/decimal"
	"go.uber.org/zap"
)

var errorStatusMap = map[error]int{
	domain.ErrInternal:        http.StatusInternalServerError,
	domain.ErrDataNotFound:    http.StatusNotFound,
	domain.ErrConflictingData: http.StatusConflict,

	domain.ErrInvalidCredentials:         http.StatusUnauthorized,
	domain.ErrUnauthorized:               http.StatusUnauthorized,
	domain.ErrEmptyAuthorizationHeader:   http.StatusUnauthorized,
	domain.ErrInvalidAuthorizationHeader: http.StatusUnauthorized,
	domain.ErrInvalidAuthorizationType:   http.StatusUnauthorized,
	domain.ErrInvalidToken:               http.StatusUnauthorized,
	domain.ErrExpiredToken:               http.StatusUnauthorized,
	domain.ErrForbidden:                  http.StatusForbidden,

	domain.ErrNoUpdatedData: http.StatusBadRequest,
	domain.ErrBadRequest:    http.StatusBadRequest,

	domain.ErrOrderAlreadyAcceptedBAnotherUser: http.StatusConflict,
	domain.ErrOrderAlreadyAcceptedByUser:       http.StatusOK,
	domain.ErrOrderBadNumber:                   http.StatusUnprocessableEntity,
	domain.ErrOrderDoubleWithdraw:              http.StatusUnprocessableEntity,
	domain.ErrInsufficientBalance:              http.StatusPaymentRequired,
}

type jsonDecimal decimal.Decimal

func (j jsonDecimal) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf("%f", decimal.Decimal(j))
	return []byte(s), nil
}

type Handler struct {
	logger *zap.Logger
}

func NewHandler(logger *zap.Logger) *Handler {
	return &Handler{logger: logger}
}

// handleValidationError sends an error response for some specific request validation error
func (h *Handler) handleValidationError(ctx *gin.Context, err error) {
	ctx.Status(http.StatusBadRequest)
}

// handleAbort sends an error response and aborts the request with the specified status code and error message
func (h *Handler) handleAbort(ctx *gin.Context, err error) {
	statusCode, ok := errorStatusMap[err]
	if !ok {
		statusCode = http.StatusInternalServerError
		h.logger.Error("aborting request", zap.Error(err))
	}
	ctx.AbortWithError(statusCode, err)
}

func (h *Handler) handleError(ctx *gin.Context, err error) {
	statusCode, ok := errorStatusMap[err]
	if !ok {
		statusCode = http.StatusInternalServerError
		h.logger.Error("error processing request", zap.Error(err))
	}
	ctx.Status(statusCode)
}

// handleSuccess sends a success response with the specified status code and optional data
func (h *Handler) handleSuccessWithStatus(ctx *gin.Context, data any, status int) {
	if data != nil {
		ctx.JSON(http.StatusOK, data)
	} else {
		ctx.Status(status)
	}
}

func (h *Handler) handleSuccess(ctx *gin.Context, data any) {
	h.handleSuccessWithStatus(ctx, data, http.StatusOK)
}
