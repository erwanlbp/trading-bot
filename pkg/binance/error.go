package binance

import (
	"errors"

	"github.com/adshao/go-binance/v2/common"
)

const (
	BinanceErrorInvalidSymbol   int64 = -1121
	BinanceErrorInvalidQuantity int64 = -1013
)

var ErrNoPriceFoundAtTime = errors.New("no_price_found_at_time")

func ErrorIs(err error, code int64) bool {
	if err == nil {
		return false
	}
	if common.IsAPIError(err) {
		if apiError, ok := err.(*common.APIError); ok {
			return apiError.Code == code
		}
	}
	return false
}
