package backtesting

import (
	"context"
	"fmt"
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
	Logger    *log.Logger
	EventBus  *eventbus.Bus
	Blacklist binance.SymbolBlackListGetter

	realClient *binance.Client

	mtx       sync.RWMutex
	balances  map[string]decimal.Decimal
	lastOrder *binance_api.Order
	prices    map[string]SymbolPrices
}

type SymbolPrices struct {
	Start, End time.Time
	Prices     map[time.Time]decimal.Decimal
}

func NewClient(l *log.Logger, cf *configfile.ConfigFile, eb *eventbus.Bus, sbg binance.SymbolBlackListGetter) *BacktestingClient {

	realClient := binance.NewClient(l, cf, eb, sbg)

	client := BacktestingClient{
		Logger:     l,
		EventBus:   eb,
		Blacklist:  sbg,
		realClient: realClient,
		balances:   make(map[string]decimal.Decimal),
		prices:     make(map[string]SymbolPrices),
	}

	return &client
}

func (c *BacktestingClient) GetBalance(ctx context.Context, coins ...string) (map[string]decimal.Decimal, error) {
	c.mtx.RLock()
	defer c.mtx.RUnlock()

	var res map[string]decimal.Decimal = make(map[string]decimal.Decimal)

	for _, coin := range coins {
		if b, ok := c.balances[coin]; ok {
			res[coin] = b
		}
	}

	return res, nil
}

func (c *BacktestingClient) GetCoinsPrice(ctx context.Context, coins, altCoins []string) (map[string]binance.CoinPrice, error) {

	var prices map[string]binance.CoinPrice = make(map[string]binance.CoinPrice)

	symbols := binance.GetSymbols(coins, altCoins, c.Blacklist)

	for _, symbol := range symbols {
		price, err := c.GetSymbolPrice(ctx, symbol)
		if err != nil {
			if binance.ErrorIs(err, binance.BinanceErrorInvalidSymbol) {
				c.EventBus.Notify(eventbus.FoundUnexistingSymbol(symbol))
				continue
			}
			return nil, err
		}
		prices[symbol] = binance.CoinPrice{
			Symbol:    symbol,
			Price:     price,
			Timestamp: util.Now(),
		}
	}

	return prices, nil
}

func (c *BacktestingClient) GetSymbolPriceAtTime(ctx context.Context, symbol string, date time.Time) (binance.CoinPrice, error) {
	symbolPrices, ok := c.prices[symbol]
	if !ok || symbolPrices.Start.After(date) || symbolPrices.End.Before(date) {
		prices, err := c.realClient.GetSymbolPricesFromTime(ctx, symbol, date.Add(6*time.Hour))
		if err != nil {
			return binance.CoinPrice{}, err
		}
		if len(prices) > 0 {
			first := prices[0]
			last := prices[len(prices)-1]
			c.Logger.Debug(fmt.Sprintf("Fetched prices of %s, from %s to %s (%d values)", symbol, first.Timestamp, last.Timestamp, len(prices)))

			pricesMap := make(map[time.Time]decimal.Decimal, len(prices))
			for _, price := range prices {
				pricesMap[price.Timestamp.Truncate(time.Minute)] = price.Price
			}

			c.prices[symbol] = SymbolPrices{
				Start:  first.Timestamp,
				End:    last.Timestamp,
				Prices: pricesMap,
			}

			symbolPrices = c.prices[symbol]
		}
	}

	price := symbolPrices.Prices[date.Truncate(time.Minute)]
	return binance.CoinPrice{
		Symbol:    symbol,
		Price:     price,
		Timestamp: date.Truncate(time.Minute),
	}, nil
}

func (c *BacktestingClient) GetSymbolPrice(ctx context.Context, symbol string) (decimal.Decimal, error) {
	price, err := c.GetSymbolPriceAtTime(ctx, symbol, util.Now())
	return price.Price, err
}

func (c *BacktestingClient) GetSymbolInfos(ctx context.Context, symbol string) (binance_api.Symbol, error) {
	return c.realClient.GetSymbolInfos(ctx, symbol)
}

// Fees are always the default in backtesting
func (c *BacktestingClient) GetFee(ctx context.Context, symbol string) (decimal.Decimal, error) {
	return binance.DefaultFee, nil
}

// Fees are always the default in backtesting
func (c *BacktestingClient) RefreshFees(ctx context.Context) {}

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
