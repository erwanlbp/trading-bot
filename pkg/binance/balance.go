package binance

import (
	"context"
	"fmt"
	"strconv"

	"github.com/erwanlbp/trading-bot/pkg/util"
)

func (c *Client) GetBalance(ctx context.Context, coins ...string) (map[string]float64, error) {
	res := make(map[string]float64)

	account, err := c.client.NewGetAccountService().Do(ctx)
	if err != nil {
		return nil, err
	}

	for _, b := range account.Balances {
		if !util.Exists(coins, func(coin string) bool { return coin == b.Asset }) {
			continue
		}
		balance, err := strconv.ParseFloat(b.Free, 64)
		if err != nil {
			return nil, fmt.Errorf("failed parsing balance for %s(%s): %w", b.Asset, b.Free, err)
		}
		res[b.Asset] = balance
	}

	return res, nil
}
