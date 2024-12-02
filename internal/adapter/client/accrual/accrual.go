package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/MikeRez0/ypgophermart/internal/adapter/config"
	"github.com/MikeRez0/ypgophermart/internal/core/domain"
	"github.com/MikeRez0/ypgophermart/internal/core/port"
	"github.com/govalues/decimal"
	"go.uber.org/zap"
)

type AccrualClient struct {
	logger     *zap.Logger
	host       string
	orderQueue chan uint64
}

func NewAccrualClient(cfg *config.Accrual, log *zap.Logger) (*AccrualClient, error) {
	return &AccrualClient{
		host:       cfg.HostString,
		logger:     log,
		orderQueue: make(chan uint64, 2),
	}, nil
}

type accrualResponse struct {
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
	Order   string  `json:"order"`
}

type errAccrualRequest struct {
	error
	NeedSleep  bool
	RetryAfter time.Duration
}

func (e *errAccrualRequest) Error() string {
	return fmt.Sprintf("Too Many Requests. Retry-After: %s", e.RetryAfter)
}

type orderAccrualStatus struct {
	Status      string
	OrderNumber uint64
	Accrual     decimal.Decimal
}

func (c *AccrualClient) ScheduleOrderAccrual(orderNumber uint64) {
	c.logger.Debug("> put order in queue (schedule)", zap.Uint64("order", orderNumber))
	c.orderQueue <- orderNumber
	c.logger.Debug("< put order in queue (schedule)", zap.Uint64("order", orderNumber))
}

func (c *AccrualClient) ScheduleAccrualService(ctx context.Context, updater port.OrderAccrualUpdater, workers int) {
	pause := sync.WaitGroup{}

	for range workers {
		go func(number chan uint64) {
			for {
				select {
				case orderNumber := <-number:

					// make request
					// wait for pause canceling
					pause.Wait()
					c.logger.Debug("Start processing order accrual",
						zap.Uint64("order", orderNumber))

					accrualStatus, err := c.requestAccrual(orderNumber)
					if err != nil {
						if e, ok := err.(*errAccrualRequest); ok {
							c.logger.Debug("Need wait for retry for order accrual",
								zap.Uint64("order", orderNumber),
								zap.Bool("NeedPause", e.NeedSleep),
								zap.Int("Retry-after", int(e.RetryAfter)))

							c.logger.Debug("Pause for requests",
								zap.Uint64("RetryAfter", uint64(e.RetryAfter)))
							r := time.NewTimer(e.RetryAfter)
							pause.Add(1)
							select {
							case <-r.C:
								c.logger.Debug("Pause finished",
									zap.Uint64("order", orderNumber))

								pause.Done()

								c.logger.Debug("> put order in queue (retry after pause)", zap.Uint64("order", orderNumber))
								c.orderQueue <- orderNumber
								c.logger.Debug("< put order in queue (retry after pause)", zap.Uint64("order", orderNumber))
							}

							continue
						}
						c.logger.Error("Unexpected error on request", zap.Error(err))
						go c.retryRequest(orderNumber, 3*time.Second)
						continue
					}

					needRetry, err := c.processOrder(context.Background(), accrualStatus, updater)

					if needRetry {
						c.logger.Error("Need retry for not finished order", zap.Error(err))
						go c.retryRequest(orderNumber, 3*time.Second)
					}

					c.logger.Debug("Finished processing order accrual",
						zap.Uint64("order", orderNumber))
				case <-ctx.Done():
					c.logger.Debug("Finished worker")
				}
			}
		}(c.orderQueue)
	}
}

func (c *AccrualClient) retryRequest(orderNumber uint64, waitFor time.Duration) {
	r := time.NewTimer(waitFor)

	select {
	case <-r.C:
		c.logger.Debug("Next request for order accrual",
			zap.Uint64("order", orderNumber))

		c.logger.Debug("> put order in queue (retry request)", zap.Uint64("order", orderNumber))
		c.orderQueue <- orderNumber
		c.logger.Debug("< put order in queue (retry request)", zap.Uint64("order", orderNumber))
	}

}

func (c *AccrualClient) requestAccrual(orderNumber uint64) (*orderAccrualStatus, error) {
	requestStr := "http://" + c.host + "/api/orders/" + strconv.Itoa(int(orderNumber))
	req, err := http.NewRequest(http.MethodGet, requestStr, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error on %s : %w", requestStr, err)
	}

	c.logger.Debug("Fire request for order accrual",
		zap.Uint64("order", orderNumber))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error %s : %w", requestStr, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusTooManyRequests {
			var retryAfter time.Duration
			sec, err := strconv.Atoi(resp.Header.Get("Retry-After"))
			if err != nil {
				retryAfter = time.Duration(10)
			} else {
				retryAfter = time.Duration(sec)
			}
			// too many requests, sleep for Retry-after seconds
			return nil, &errAccrualRequest{RetryAfter: retryAfter * time.Second, NeedSleep: true}
		} else if resp.StatusCode == http.StatusNoContent {
			// Order not registered, wait 10 sec
			return nil, &errAccrualRequest{RetryAfter: 10 * time.Second}
		}
		c.logger.Error("unexpected status for request",
			zap.Uint64("order", orderNumber), zap.Int("status", resp.StatusCode))
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

	return &orderAccrualStatus{
		OrderNumber: uint64(num),
		Accrual:     acc,
		Status:      result.Status,
	}, nil
}

func (c *AccrualClient) processOrder(ctx context.Context, status *orderAccrualStatus,
	orderAccrualUpdater port.OrderAccrualUpdater) (bool, error) {
	if status.Status == "PROCESSED" {
		err := orderAccrualUpdater.AccrualOrder(ctx, status.OrderNumber, status.Accrual)
		if err != nil {
			c.logger.Error("accrual order error", zap.Error(err))
			return true, err
		}
		return false, nil
	}

	if status.Status == "INVALID" {
		return false, orderAccrualUpdater.UpdateOrderStatus(ctx, status.OrderNumber, domain.OrderStatusInvalid)
	} else {
		return true, orderAccrualUpdater.UpdateOrderStatus(ctx, status.OrderNumber, domain.OrderStatusProcessing)
	}
}
