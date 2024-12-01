package http

import (
	"github.com/MikeRez0/ypgophermart/internal/core/domain"
	"github.com/MikeRez0/ypgophermart/internal/core/port"
	"github.com/MikeRez0/ypgophermart/internal/core/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct {
	Handler
	service port.Service
}

type UserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func NewUserHandler(service port.Service, logger *zap.Logger) (*UserHandler, error) {
	return &UserHandler{
		Handler: Handler{logger: logger},
		service: service}, nil
}

func (uh *UserHandler) RegisterUser(ctx *gin.Context) {
	userReq := UserRequest{}
	err := ctx.ShouldBindBodyWithJSON(&userReq)
	if err != nil {
		uh.handleValidationError(ctx, err)
		return
	}

	// Hash password
	hashed, err := utils.HashPassword(userReq.Password)
	if err != nil {
		uh.handleError(ctx, domain.ErrInternal)
		return
	}
	userReq.Password = hashed

	user := &domain.User{
		Login:    userReq.Login,
		Password: userReq.Password,
	}

	_, err = uh.service.RegisterUser(ctx, user)
	if err != nil {
		uh.handleError(ctx, err)
		return
	}

	// Token return
	uh.LoginUser(ctx)
}

func (uh *UserHandler) LoginUser(ctx *gin.Context) {
	userReq := UserRequest{}
	err := ctx.ShouldBindBodyWithJSON(&userReq)
	if err != nil {
		uh.handleValidationError(ctx, err)
		return
	}

	token, err := uh.service.LoginUser(ctx, userReq.Login, userReq.Password)
	if err != nil {
		uh.handleError(ctx, err)
		return
	}

	uh.handleSuccess(ctx, struct {
		Token string `json:"token"`
	}{Token: token})
}
