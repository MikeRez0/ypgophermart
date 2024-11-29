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

type Order struct {
	UserID     uint64
	Number     uint64
	Accrual    decimal.Decimal
	Status     OrderStatus
	UploadedAt time.Time
	User       *User
}
