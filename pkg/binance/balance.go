package binance

import (
	"context"
	"fmt"

	"github.com/erwanlbp/trading-bot/pkg/util"
	"github.com/shopspring/decimal"
)

func (c *Client) GetBalance(ctx context.Context, coins ...string) (map[string]decimal.Decimal, error) {

	account, err := c.client.NewGetAccountService().Do(ctx)
	if err != nil {
		return nil, err
	}

	res := make(map[string]decimal.Decimal)
	for _, b := range account.Balances {
		if len(coins) > 0 && !util.Exists(coins, func(coin string) bool { return coin == b.Asset }) {
			continue
		}
		balance, err := decimal.NewFromString(b.Free)
		if err != nil {
			return nil, fmt.Errorf("failed parsing balance for %s(%s): %w", b.Asset, b.Free, err)
		}
		res[b.Asset] = balance
	}

	return res, nil
}
