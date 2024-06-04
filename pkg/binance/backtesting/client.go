package backtesting

import (
	"context"
	"fmt"
	"sync"
	"time"

	binance_api "github.com/adshao/go-binance/v2"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

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

	KlineInterval string
	TimeToAdd     time.Duration

	mtx       sync.RWMutex
	balances  map[string]decimal.Decimal
	lastOrder *binance_api.Order
	prices    map[string]SymbolPrices
}

type SymbolPrices struct {
	Start, End time.Time
	Prices     map[time.Time]decimal.Decimal
}

func NewClient(l *log.Logger, cf *configfile.ConfigFile, eb *eventbus.Bus, sbg binance.SymbolBlackListGetter, initialBalance decimal.Decimal, klineInterval string) *BacktestingClient {

	realClient := binance.NewClient(l, cf, eb, sbg)

	klineIntervalDuration, err := time.ParseDuration(klineInterval)
	if err != nil {
		panic(fmt.Errorf("cannot parse kline interval '%s' to duration: %w", klineInterval, err))
	}

	client := BacktestingClient{
		Logger:        l,
		EventBus:      eb,
		Blacklist:     sbg,
		realClient:    realClient,
		KlineInterval: klineInterval,
		TimeToAdd:     klineIntervalDuration * 450,
		balances:      make(map[string]decimal.Decimal),
		prices:        make(map[string]SymbolPrices),
	}

	client.balances[cf.Bridge] = initialBalance

	return &client
}

func (c *BacktestingClient) GetBalance(ctx context.Context, coins ...string) (map[string]decimal.Decimal, error) {
	c.mtx.RLock()
	defer c.mtx.RUnlock()

	if len(coins) == 0 {
		return c.balances, nil
	}

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
		prices, err := c.realClient.GetSymbolPricesFromTime(ctx, symbol, date.Add(c.TimeToAdd), c.KlineInterval)
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

	price, ok := symbolPrices.Prices[date.Truncate(time.Minute)]
	if !ok {
		util.DebugPrintJson(symbolPrices.Prices[date.Truncate(time.Minute)])
		return binance.CoinPrice{}, fmt.Errorf("cannot find price")
	}

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
	return c.realClient.GetFee(ctx, symbol)
}

// Fees are always the default in backtesting
func (c *BacktestingClient) RefreshFees(ctx context.Context) {
	c.realClient.RefreshFees(ctx)
}

func (c *BacktestingClient) Sell(ctx context.Context, coin, stableCoin string) (binance.OrderResult, error) {
	var res binance.OrderResult

	balances, err := c.GetBalance(ctx)
	if err != nil {
		return res, fmt.Errorf("failed to get balances: %w", err)
	}

	symbol := util.Symbol(coin, stableCoin)

	price, err := c.GetSymbolPrice(ctx, symbol)
	if err != nil {
		return res, fmt.Errorf("failed to get symbol price: %w", err)
	}

	symbolInfo, err := c.GetSymbolInfos(ctx, symbol)
	if err != nil {
		return res, fmt.Errorf("failed to get symbol infos: %w", err)
	}

	// Calculate how much quantity we can sell
	step := symbolInfo.LotSizeFilter().StepSize

	fee, err := c.GetFee(ctx, symbol)
	if err != nil {
		return res, fmt.Errorf("failed fetching fee: %w", err)
	}

	quantity := decimal.RequireFromString(binance.StepSizeFormat(balances[coin], step))
	quantityWithoutFee := quantity.Mul(decimal.NewFromInt(1).Sub(fee))

	c.Logger.Info(fmt.Sprintf("I have %s %s and %s %s. I'll sell %s %s, at price %s", balances[coin], coin, balances[stableCoin], stableCoin, quantity, coin, price), zap.String("step", step))

	c.balances[coin] = c.balances[coin].Sub(quantity)
	c.balances[stableCoin] = c.balances[stableCoin].Add(quantityWithoutFee.Mul(price))

	return binance.OrderResult{
		Order: &binance_api.Order{
			Status:           binance_api.OrderStatusTypeFilled,
			Price:            price.String(),
			ExecutedQuantity: quantity.String(),
			Time:             util.Now().UnixMilli(),
		},
	}, nil
}

func (c *BacktestingClient) Buy(ctx context.Context, coin, stableCoin string) (binance.OrderResult, error) {
	var res binance.OrderResult

	balances, err := c.GetBalance(ctx, coin, stableCoin)
	if err != nil {
		return res, fmt.Errorf("failed to get balances: %w", err)
	}

	symbol := util.Symbol(coin, stableCoin)

	price, err := c.GetSymbolPrice(ctx, symbol)
	if err != nil {
		return res, fmt.Errorf("failed to get symbol price: %w", err)
	}

	symbolInfo, err := c.GetSymbolInfos(ctx, symbol)
	if err != nil {
		return res, fmt.Errorf("failed to get symbol infos: %w", err)
	}

	// Calculate how much quantity we can sell
	step := symbolInfo.LotSizeFilter().StepSize

	fee, err := c.GetFee(ctx, symbol)
	if err != nil {
		return res, fmt.Errorf("failed fetching fee: %w", err)
	}

	quantity := decimal.RequireFromString(binance.StepSizeFormat(balances[stableCoin].Div(price), step))
	quantityWithoutFee := quantity.Mul(decimal.NewFromInt(1).Sub(fee))

	c.Logger.Info(fmt.Sprintf("I have %s %s and %s %s. I'll buy %s %s, at price %s", balances[coin], coin, balances[stableCoin], stableCoin, quantity, coin, price), zap.String("step", step))

	c.balances[stableCoin] = c.balances[stableCoin].Sub(quantity.Mul(price))
	c.balances[coin] = c.balances[coin].Add(quantityWithoutFee)

	return binance.OrderResult{
		Order: &binance_api.Order{
			Status:           binance_api.OrderStatusTypeFilled,
			Price:            price.String(),
			ExecutedQuantity: quantity.String(),
			Time:             util.Now().UnixMilli(),
		},
	}, nil
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

func (c *BacktestingClient) LogBalances(ctx context.Context) {
	b, err := c.GetBalance(ctx)
	if err != nil {
		c.Logger.Error("Failed to get balances", zap.Error(err))
	} else {
		c.Logger.Info(fmt.Sprintf("Balances are %s", util.ToJSON(b)))
	}
}
