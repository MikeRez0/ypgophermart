package accrual

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/MikeRez0/ypgophermart/internal/adapter/config"
	"github.com/MikeRez0/ypgophermart/internal/core/port"
	"github.com/govalues/decimal"
	"go.uber.org/zap"
)

type AccrualClient struct {
	logger *zap.Logger
	host   string
}

func NewAccrualClient(cfg *config.Accrual, log *zap.Logger) (*AccrualClient, error) {
	return &AccrualClient{
		host:   cfg.HostString,
		logger: log,
	}, nil
}

type accrualResponse struct {
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
	Order   string  `json:"order"`
}

func (c *AccrualClient) GetOrderAccrual(orderNumber uint64) (*port.OrderAccrualResonse, error) {
	requestStr := "http://" + c.host + "/api/orders/" + strconv.Itoa(int(orderNumber))
	req, err := http.NewRequest(http.MethodGet, requestStr, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error on %s : %w", requestStr, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error %s : %w", requestStr, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response %v for request %s", resp.StatusCode, requestStr)
	}

	var result accrualResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("error on response decode: %w", err)
	}
	num, err := strconv.Atoi(result.Order)
	if err != nil {
		return nil, fmt.Errorf("error on response decode: %w", err)
	}
	acc, err := decimal.NewFromFloat64(result.Accrual)
	if err != nil {
		return nil, fmt.Errorf("error on response decode: %w", err)
	}

	return &port.OrderAccrualResonse{
		OrderNumber: uint64(num),
		Accrual:     acc,
		Status:      result.Status,
	}, nil
}
