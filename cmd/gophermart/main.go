package main

import (
	"context"

	"github.com/MikeRez0/ypgophermart/internal/adapter/auth"
	"github.com/MikeRez0/ypgophermart/internal/adapter/config"
	"github.com/MikeRez0/ypgophermart/internal/adapter/handler/http"
	"github.com/MikeRez0/ypgophermart/internal/adapter/storage"
	"github.com/MikeRez0/ypgophermart/internal/adapter/storage/repository"
	"github.com/MikeRez0/ypgophermart/internal/core/service"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	conf, err := config.NewConfig()
	if err != nil {
		logger.Fatal("config error", zap.Error(err))
	}

	ctx := context.Background()

	db, err := storage.NewDBStorage(ctx, conf.Database)
	if err != nil {
		logger.Fatal("database error", zap.Error(err))
	}
	err = db.RunMigrations()
	if err != nil {
		logger.Fatal("database migration error", zap.Error(err))
	}

	repo, err := repository.NewRepository(db)
	if err != nil {
		logger.Fatal("order repo creating error", zap.Error(err))
	}
	tokenService, err := auth.New()
	if err != nil {
		logger.Fatal("token service creating error", zap.Error(err))
	}

	svc, err := service.NewService(repo, tokenService)
	if err != nil {
		logger.Fatal("order service creating error", zap.Error(err))
	}

	orderHandler, err := http.NewOrderHandler(svc)
	if err != nil {
		logger.Fatal("order handler creating error", zap.Error(err))
	}
	userHandler, err := http.NewUserHandler(svc)
	if err != nil {
		logger.Fatal("user handler creating error", zap.Error(err))
	}

	r, err := http.NewRouter(conf.HTTP, tokenService, orderHandler, userHandler)
	if err != nil {
		logger.Fatal("router creating error", zap.Error(err))
	}

	err = r.Serve(conf.HTTP.HostString)
	if err != nil {
		logger.Fatal("router start error", zap.Error(err))
	}
}
