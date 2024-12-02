package domain

import (
	"time"

	"github.com/govalues/decimal"
)

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
	OrderStatusInvalid    OrderStatus = "INVALID"
)

type OrderNumber string

type Order struct {
	Number     OrderNumber
	UserID     uint64
	Accrual    decimal.Decimal
	Withdrawal decimal.Decimal
	Status     OrderStatus
	UploadedAt time.Time
}
