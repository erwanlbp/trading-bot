package binance

import (
	"github.com/adshao/go-binance/v2/common"
)

const (
	BinanceErrorInvalidSymbol int64 = -1121
)

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
