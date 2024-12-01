package port

import "github.com/govalues/decimal"

type OrderAccrualResonse struct {
	Status      string
	OrderNumber uint64
	Accrual     decimal.Decimal
}

type AccrualServiceClient interface {
	GetOrderAccrual(orderNumber uint64) (*OrderAccrualResonse, error)
}
