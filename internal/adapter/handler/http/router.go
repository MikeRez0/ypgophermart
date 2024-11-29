package http

import (
	"github.com/MikeRez0/ypgophermart/internal/adapter/config"
	"github.com/MikeRez0/ypgophermart/internal/core/port"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Router struct {
	*gin.Engine
}

func NewRouter(
	conf *config.HTTP,
	tokenService port.TokenService,
	orderHandler *OrderHandler,
	userHandler *UserHandler) (*Router, error) {

	router := gin.New()

	// Swagger
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api")
	{
		user := api.Group("/user")
		{
			user.POST("/register", userHandler.RegisterUser)
			user.POST("/login", userHandler.LoginUser)

			orders := user.Group("/orders")
			{
				orders.Use(authCheck(tokenService))
				orders.POST("/:order", orderHandler.CreateOrder)
				orders.GET("", orderHandler.ListOrdersByUser)
			}
		}
	}

	return &Router{router}, nil
}

// Serve starts the HTTP server
func (r *Router) Serve(listenAddr string) error {
	return r.Run(listenAddr)
}
