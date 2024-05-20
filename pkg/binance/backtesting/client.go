package backtesting

import (
	"context"
	"sync"
	"time"

	binance_api "github.com/adshao/go-binance/v2"
	"github.com/shopspring/decimal"

	"github.com/erwanlbp/trading-bot/pkg/binance"
	"github.com/erwanlbp/trading-bot/pkg/config/configfile"
	"github.com/erwanlbp/trading-bot/pkg/eventbus"
	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

type BacktestingClient struct {
	realClient binance.Client

	nowFunc util.NowFunc

	mtx       sync.RWMutex
	balances  map[string]decimal.Decimal
	lastOrder *binance_api.Order
}

func NewClient(nowFunc util.NowFunc, l *log.Logger, cf *configfile.ConfigFile, eb *eventbus.Bus, sbg binance.SymbolBlackListGetter) *BacktestingClient {

	realClient := binance.NewClient(l, cf, eb, sbg)

	client := BacktestingClient{
		realClient: realClient,
		nowFunc:    nowFunc,
	}

	return &client
}

func (c *BacktestingClient) GetBalance(ctx context.Context, coins ...string) (map[string]decimal.Decimal, error) {
	c.mtx.RLock()
	defer c.mtx.RUnlock()

	var res map[string]decimal.Decimal

	for _, coin := range coins {
		if b, ok := c.balances[coin]; ok {
			res[coin] = b
		}
	}

	return res, nil
}

func (c *BacktestingClient) GetCoinsPrice(ctx context.Context, coins, altCoins []string) (map[string]binance.CoinPrice, error) {
	// TODO Use GetSymbolPriceAtTime with NowFunc ?
	panic("not implemented")
	return nil, nil
}

func (c *BacktestingClient) GetSymbolPriceAtTime(ctx context.Context, symbol string, date time.Time) (binance.CoinPrice, error) {
	return c.realClient.GetSymbolPriceAtTime(ctx, symbol, date)
}

func (c *BacktestingClient) GetSymbolPrice(ctx context.Context, symbol string) (decimal.Decimal, error) {
	price, err := c.GetSymbolPriceAtTime(ctx, symbol, c.nowFunc())
	return price.Price, err
}

func (c *BacktestingClient) dichotomicPriceFetching(ctx context.Context, symbols []string) ([]*binance_api.SymbolPrice, error) {
	// TODO Make a helper of this func in the real client
	panic("not implemented")
}

func (c *BacktestingClient) GetSymbolInfos(ctx context.Context, symbol string) (binance_api.Symbol, error) {
	return c.realClient.GetSymbolInfos(ctx, symbol)
}

// Fees are always the default in backtesting
func (c *BacktestingClient) GetFee(ctx context.Context, symbol string) (decimal.Decimal, error) {
	return binance.DefaultFee, nil
}

// Fees are always the default in backtesting
func (c *BacktestingClient) RefreshFees(ctx context.Context) {
	return
}

func (c *BacktestingClient) Sell(ctx context.Context, coin, stableCoin string) (binance.OrderResult, error) {
	panic("not implemented")
}

func (c *BacktestingClient) Buy(ctx context.Context, coin, stableCoin string) (binance.OrderResult, error) {
	panic("not implemented")
}

func (c *BacktestingClient) TradeLock() (func(), error) {
	return c.realClient.TradeLock()
}

func (c *BacktestingClient) IsTradeInProgress() bool {
	return c.realClient.IsTradeInProgress()
}

func (c *BacktestingClient) WaitForOrderCompletion(ctx context.Context, symbol string, orderId int64) (binance.OrderResult, error) {
	return binance.OrderResult{
		Order: c.lastOrder,
	}, nil
}

func (c *BacktestingClient) LogBalances(ctx context.Context) {}
