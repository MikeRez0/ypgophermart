package logger

import (
	"github.com/MikeRez0/ypgophermart/internal/adapter/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(conf *config.App) *zap.Logger {
	lvl, err := zap.ParseAtomicLevel(conf.LogLevel)
	if err != nil {
		zap.L().Error("error parsing log level", zap.Error(err))
		return nil
	}

	if conf.Mode == config.AppModeDevelop {
		cfg := zap.NewDevelopmentConfig()
		cfg.Level = lvl
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

		logger := zap.Must(cfg.Build())

		return logger
	} else {
		cfg := zap.NewProductionConfig()
		cfg.Level = lvl
		logger := zap.Must(cfg.Build())

		return logger
	}

}
