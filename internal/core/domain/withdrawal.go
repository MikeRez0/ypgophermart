package domain

import (
	"time"

	"github.com/govalues/decimal"
)

type Withdrawal struct {
	OrderID     uint64
	Sum         decimal.Decimal
	ProcessedAt time.Time
	Order       *Order
}
