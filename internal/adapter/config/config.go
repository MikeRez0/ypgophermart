package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Database *Database
	HTTP     *HTTP
	Accrual  *Accrual
	App      *App
}

const AppModeProduction = "PROD"
const AppModeDevelop = "DEV"

type App struct {
	LogLevel string `env:"LOG_LEVEL"`
	Mode     string
}

type Database struct {
	DSN string `env:"DATABASE_URI"`
}

type HTTP struct {
	HostString string `env:"RUN_ADDRESS"`
}

type Accrual struct {
	HostString string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func NewConfig() (*Config, error) {
	var db Database
	var http HTTP
	var accrual Accrual
	var app App

	flag.StringVar(&db.DSN, "d", "", "Database string")
	flag.StringVar(&http.HostString, "a", `localhost:8080`, "HTTP server endpoint")
	flag.StringVar(&accrual.HostString, "r", "", "Accrual system address")
	flag.StringVar(&app.LogLevel, "l", `error`, "Log level")
	flag.StringVar(&app.Mode, "m", `DEV`, "PROD / DEV")
	flag.Parse()

	err := env.Parse(&db)
	if err != nil {
		return nil, fmt.Errorf("error parsing env database config: %w", err)
	}
	err = env.Parse(&http)
	if err != nil {
		return nil, fmt.Errorf("error parsing http config: %w", err)
	}
	err = env.Parse(&app)
	if err != nil {
		return nil, fmt.Errorf("error parsing app config: %w", err)
	}
	err = env.Parse(&accrual)
	if err != nil {
		return nil, fmt.Errorf("error parsing accrual config: %w", err)
	}

	config := Config{
		Database: &db,
		HTTP:     &http,
		Accrual:  &accrual,
		App:      &app,
	}

	return &config, nil
}
