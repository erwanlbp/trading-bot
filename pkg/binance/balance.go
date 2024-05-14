package binance

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/erwanlbp/trading-bot/pkg/util"
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
		if !balance.Equal(decimal.Zero) {
			res[b.Asset] = balance
		}
	}

	return res, nil
}

func (c *Client) GetBalanceValue(ctx context.Context, altCoins []string) (map[string]decimal.Decimal, error) {
	balances, err := c.GetBalance(ctx, append(c.ConfigFile.Coins, c.ConfigFile.Bridge)...)
	if err != nil {
		return nil, err
	}

	var balancePositiveCoin []string
	for s, d := range balances {
		if d.GreaterThan(decimal.Zero) {
			balancePositiveCoin = append(balancePositiveCoin, s)
		}
	}

	prices, err := c.GetCoinsPrice(ctx, balancePositiveCoin, altCoins)
	if err != nil {
		return nil, err
	}

	res := map[string]decimal.Decimal{}
	for _, price := range prices {
		res[price.AltCoin] = res[price.AltCoin].Add(price.Price.Mul(balances[price.Coin]))
	}

	return res, nil
}
