package util

import "github.com/shopspring/decimal"

func Float64(d decimal.Decimal) float64 {
	v, _ := d.Float64()
	return v
}
