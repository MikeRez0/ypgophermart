package http

import (
	"strings"

	"github.com/MikeRez0/ypgophermart/internal/core/domain"
	"github.com/MikeRez0/ypgophermart/internal/core/port"
	"github.com/gin-gonic/gin"
)

const authHeaderKey = "Authorization"
const authType = "Bearer"
const userPayloadKey = "user_payload"

func authCheck(tokenService port.TokenService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := ctx.Request.Header.Get(authHeaderKey)
		if len(header) == 0 {
			handleAbort(ctx, domain.ErrEmptyAuthorizationHeader)
			return
		}

		words := strings.Split(header, " ")
		if len(words) != 2 {
			handleAbort(ctx, domain.ErrInvalidAuthorizationHeader)
			return
		}
		if words[0] != authType {
			handleAbort(ctx, domain.ErrInvalidAuthorizationType)
			return
		}
		token := words[1]
		payload, err := tokenService.VerifyToken(token)
		if err != nil {
			handleAbort(ctx, domain.ErrInvalidToken)
			return
		}

		ctx.Set(userPayloadKey, payload)

		ctx.Next()
	}
}

func getAuthPayload(ctx *gin.Context) *port.TokenPayload {
	return ctx.MustGet(userPayloadKey).(*port.TokenPayload)
}
