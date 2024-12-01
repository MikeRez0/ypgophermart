package http

import (
	"strings"
	"time"

	"github.com/MikeRez0/ypgophermart/internal/core/domain"
	"github.com/MikeRez0/ypgophermart/internal/core/port"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const authHeaderKey = "Authorization"
const authType = "Bearer"
const userPayloadKey = "user_payload"

func authCheck(tokenService port.TokenService, logger *zap.Logger) gin.HandlerFunc {
	authHandler := &Handler{logger: logger}
	return func(ctx *gin.Context) {
		header := ctx.Request.Header.Get(authHeaderKey)
		if len(header) == 0 {
			authHandler.handleAbort(ctx, domain.ErrEmptyAuthorizationHeader)
			return
		}

		words := strings.Split(header, " ")
		if len(words) != 2 {
			authHandler.handleAbort(ctx, domain.ErrInvalidAuthorizationHeader)
			return
		}
		if words[0] != authType {
			authHandler.handleAbort(ctx, domain.ErrInvalidAuthorizationType)
			return
		}
		token := words[1]
		payload, err := tokenService.VerifyToken(token)
		if err != nil {
			authHandler.handleAbort(ctx, domain.ErrInvalidToken)
			return
		}

		ctx.Set(userPayloadKey, payload)

		ctx.Next()
	}
}

func getAuthPayload(ctx *gin.Context) *port.TokenPayload {
	return ctx.MustGet(userPayloadKey).(*port.TokenPayload)
}

func logRequest(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		c.Next()

		log.Info("Income HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.RequestURI),
			zap.Int("status", c.Writer.Status()),
			zap.Int("size", c.Writer.Size()),
			zap.String("duration", time.Since(t).String()))
	}
}
