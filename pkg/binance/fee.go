package binance

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/erwanlbp/trading-bot/pkg/util"
)

var InvalidFeeValue error = errors.New("can't find symbol fee")

type feesCache struct {
	mtx  sync.RWMutex
	fees map[string]decimal.Decimal
}

var allFees feesCache

func (c *Client) GetFee(ctx context.Context, symbol string) (decimal.Decimal, error) {
	allFees.mtx.RLock()
	defer allFees.mtx.RUnlock()

	// For the edge case where we could call here before the fees are initialized
	if allFees.fees == nil {
		return decimal.Zero, InvalidFeeValue
	}

	fee, ok := allFees.fees[symbol]
	if !ok {
		return decimal.Zero, InvalidFeeValue
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
		allFees.fees = make(map[string]decimal.Decimal)
	}
	for _, fee := range feeDetails {
		// TODO Not 100% sure we're maker, or not always? see https://www.binance.com/fr/support/faq/que-sont-les-makers-et-les-takers-360007720071
		feeStr := fee.MakerCommission

		feeValue, err := decimal.NewFromString(feeStr)
		if err != nil {
			c.Logger.Error("Failed parsing fee value, ignoring", zap.Error(err))
			continue
		}

		allFees.fees[fee.Symbol] = feeValue
	}
}

var DefaultFee = decimal.NewFromFloat(0.998001)

func (c *Client) GetJumpFeeMultiplier(ctx context.Context, fromCoin, toCoin, bridge string) (decimal.Decimal, error) {
	sellingFeePct, err := c.GetFee(ctx, util.Symbol(fromCoin, bridge))
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get selling fee: %w", err)
	}
	buyingFeePct, err := c.GetFee(ctx, util.Symbol(toCoin, bridge))
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get buying fee: %w", err)
	}
	return decimal.NewFromInt(1).Sub(sellingFeePct.Add(buyingFeePct).Sub(sellingFeePct.Mul(buyingFeePct))), nil

}
