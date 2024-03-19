package binance

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"go.uber.org/zap"
)

var InvalidFeeValue error = errors.New("can't find symbol fee")

type feesCache struct {
	mtx  sync.RWMutex
	fees map[string]float64
}

var allFees feesCache

func (c *Client) GetFee(ctx context.Context, symbol string) (float64, error) {
	allFees.mtx.RLock()
	defer allFees.mtx.RUnlock()

	// For the edge case where we could call here before the fees are initialized
	if allFees.fees == nil {
		return 0, InvalidFeeValue
	}

	fee, ok := allFees.fees[symbol]
	if !ok {
		return 0, InvalidFeeValue
	}
	return fee, nil
}

func (c *Client) RefreshFees(ctx context.Context) {
	allFees.mtx.Lock()
	defer allFees.mtx.Unlock()

	feeDetails, err := c.client.NewTradeFeeService().Do(ctx)
	if err != nil {
		c.Logger.Error("Failed refreshing fees", zap.Error(err))
		return
	}

	if allFees.fees == nil {
		allFees.fees = make(map[string]float64)
	}
	for _, fee := range feeDetails {
		// TODO Not 100% sure we're maker, or not always? see https://www.binance.com/fr/support/faq/que-sont-les-makers-et-les-takers-360007720071
		feeStr := fee.MakerCommission

		feeValue, err := strconv.ParseFloat(feeStr, 64)
		if err != nil {
			c.Logger.Error("Failed parsing fee value, ignoring", zap.Error(err))
			continue
		}

		allFees.fees[fee.Symbol] = feeValue
	}
}
