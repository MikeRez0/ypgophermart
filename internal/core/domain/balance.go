package domain

import "github.com/govalues/decimal"

type Balance struct {
	UserID    uint64
	Current   decimal.Decimal
	Withdrawn decimal.Decimal
}
