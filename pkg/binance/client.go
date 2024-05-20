package binance

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/erwanlbp/trading-bot/pkg/config/configfile"
	"github.com/erwanlbp/trading-bot/pkg/eventbus"
	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/refresher"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

type SymbolBlackListGetter interface {
	IsSymbolBlacklisted(symbol string) bool
}

type client struct {
	client          *binance.Client
	Logger          *log.Logger
	ConfigFile      *configfile.ConfigFile
	EventBus        *eventbus.Bus
	SymbolBlackList SymbolBlackListGetter

	tradeInProgress atomic.Bool

	coinInfosRefresher *refresher.Refresher[map[string]binance.Symbol]
}

type Client interface {
	GetBalance(ctx context.Context, coins ...string) (map[string]decimal.Decimal, error)
	GetCoinsPrice(ctx context.Context, coins, altCoins []string) (map[string]CoinPrice, error)
	GetSymbolPriceAtTime(ctx context.Context, symbol string, date time.Time) (CoinPrice, error)
	GetSymbolPrice(ctx context.Context, symbol string) (decimal.Decimal, error)
	GetSymbolInfos(ctx context.Context, symbol string) (binance.Symbol, error)
	GetFee(ctx context.Context, symbol string) (decimal.Decimal, error)
	RefreshFees(ctx context.Context)
	Sell(ctx context.Context, coin, stableCoin string) (OrderResult, error)
	Buy(ctx context.Context, coin, stableCoin string) (OrderResult, error)
	TradeLock() (func(), error)
	IsTradeInProgress() bool
	WaitForOrderCompletion(ctx context.Context, symbol string, orderId int64) (OrderResult, error)
	LogBalances(ctx context.Context)
}

func NewClient(l *log.Logger, cf *configfile.ConfigFile, eb *eventbus.Bus, sbg SymbolBlackListGetter) *client {
	if cf.TestMode {
		l.Info("Activating Binance test mode")
		binance.UseTestnet = true
	}

	client := client{
		client:          binance.NewClient(cf.Binance.APIKey, cf.Binance.APIKeySecret),
		Logger:          l,
		ConfigFile:      cf,
		EventBus:        eb,
		SymbolBlackList: sbg,
	}

	client.coinInfosRefresher = refresher.NewRefresher(l, 5*time.Minute, client.RefreshSymbolInfos, refresher.OnErrorLog(client.Logger))

	return &client
}

func (c *client) LogBalances(ctx context.Context) {
	b, err := c.GetBalance(ctx, append(c.ConfigFile.Coins, c.ConfigFile.Bridge)...)
	if err != nil {
		c.Logger.Error("Failed to get balances", zap.Error(err))
	} else {
		c.Logger.Info(fmt.Sprintf("Balances are %s", util.ToJSON(b)))
	}
}
