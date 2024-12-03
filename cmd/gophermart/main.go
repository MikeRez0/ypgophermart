package main

import (
	"context"
	"fmt"

	"github.com/MikeRez0/ypgophermart/internal/adapter/auth"
	"github.com/MikeRez0/ypgophermart/internal/adapter/client/accrual"
	"github.com/MikeRez0/ypgophermart/internal/adapter/config"
	"github.com/MikeRez0/ypgophermart/internal/adapter/handler/http"
	"github.com/MikeRez0/ypgophermart/internal/adapter/logger"
	"github.com/MikeRez0/ypgophermart/internal/adapter/storage"
	"github.com/MikeRez0/ypgophermart/internal/adapter/storage/repository"
	"github.com/MikeRez0/ypgophermart/internal/core/service"
	"go.uber.org/zap"
)

func main() {
	conf, err := config.NewConfig()
	if err != nil {
		fmt.Printf("config error:%s", err)
		return
	}

	log := logger.NewLogger(conf.App)
	if log == nil {
		fmt.Printf("error creating log")
		return
	}
	defer func() {
		err := log.Sync()
		if err != nil {
			fmt.Printf("log error: %s", err)
		}
	}()

	ctx := context.Background()

	db, err := storage.NewDBStorage(ctx, conf.Database)
	if err != nil {
		log.Error("database error", zap.Error(err))
		return
	}
	err = db.RunMigrations()
	if err != nil {
		log.Error("database migration error", zap.Error(err))
		return
	}

	repo, err := repository.NewRepository(db)
	if err != nil {
		log.Error("order repo creating error", zap.Error(err))
		return
	}
	tokenService, err := auth.New()
	if err != nil {
		log.Error("token service creating error", zap.Error(err))
		return
	}

	accrual, err := accrual.NewAccrualClient(conf.Accrual, log.Named("Accrual"))
	if err != nil {
		log.Error("accrual client creating error", zap.Error(err))
		return
	}

	svc, err := service.NewService(repo, tokenService, accrual, log.Named("Service"))
	if err != nil {
		log.Error("order service creating error", zap.Error(err))
		return
	}

	accrual.ScheduleAccrualService(ctx, svc, 5)

	userHandler, err := http.NewUserHandler(svc, log.Named("User handler"))
	if err != nil {
		log.Error("user handler creating error", zap.Error(err))
		return
	}
	orderHandler, err := http.NewOrderHandler(svc, log.Named("Order handler"))
	if err != nil {
		log.Error("order handler creating error", zap.Error(err))
		return
	}
	balanceHandler, err := http.NewBalanceHandler(svc, log)
	if err != nil {
		log.Error("balance handler creating error", zap.Error(err))
		return
	}

	r, err := http.NewRouter(conf.HTTP, tokenService, orderHandler, userHandler, balanceHandler, log.Named("Router"))
	if err != nil {
		log.Error("router creating error", zap.Error(err))
		return
	}

	err = r.Serve(conf.HTTP.HostString)
	if err != nil {
		log.Error("router serve error", zap.Error(err))
		return
	}
}
