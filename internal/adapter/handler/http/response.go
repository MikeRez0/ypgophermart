package http

import (
	"net/http"

	"github.com/MikeRez0/ypgophermart/internal/core/domain"
	"github.com/gin-gonic/gin"
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
}

// handleValidationError sends an error response for some specific request validation error
func handleValidationError(ctx *gin.Context, err error) {
	ctx.Status(http.StatusBadRequest)
}

// handleAbort sends an error response and aborts the request with the specified status code and error message
func handleAbort(ctx *gin.Context, err error) {
	statusCode, ok := errorStatusMap[err]
	if !ok {
		statusCode = http.StatusInternalServerError
	}
	ctx.AbortWithError(statusCode, err)
}

func handleError(ctx *gin.Context, err error) {
	statusCode, ok := errorStatusMap[err]
	if !ok {
		statusCode = http.StatusInternalServerError
	}
	ctx.Status(statusCode)
}

// handleSuccess sends a success response with the specified status code and optional data
func handleSuccessWithStatus(ctx *gin.Context, data any, status int) {

	if data != nil {
		ctx.JSON(http.StatusOK, data)
	} else {
		ctx.Status(status)
	}
}

func handleSuccess(ctx *gin.Context, data any) {
	handleSuccessWithStatus(ctx, data, http.StatusOK)
}
