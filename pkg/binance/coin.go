package binance

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/erwanlbp/trading-bot/pkg/util"
)

type CoinPrice struct {
	Coin      string
	AltCoin   string
	Price     float64
	Timestamp time.Time
}

func (c *Client) GetCoinsPrice(ctx context.Context, coins []string, altCoin string) ([]CoinPrice, error) {

	symbols := util.Map(coins, func(c string) string { return util.Symbol(c, altCoin) })

	prices, err := c.client.NewListPricesService().Symbols(symbols).Do(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()

	var res []CoinPrice
	for _, price := range prices {
		coin, altCoin, err := util.Unsymbol(price.Symbol, coins, altCoin)
		if err != nil {
			return nil, fmt.Errorf("couldn't unsymbol %s: %w", price.Symbol, err)
		}
		p, err := strconv.ParseFloat(price.Price, 64)
		if err != nil {
			return nil, fmt.Errorf("failed parsing price for %s(%s): %w", price.Symbol, price.Price, err)
		}
		res = append(res, CoinPrice{
			Coin:      coin,
			AltCoin:   altCoin,
			Price:     p,
			Timestamp: now,
		})
	}

	return res, nil
}
