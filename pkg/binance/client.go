package binance

import (
	"context"
	"fmt"
	"time"

	"github.com/adshao/go-binance/v2"
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

type Client struct {
	client          *binance.Client
	Logger          *log.Logger
	ConfigFile      *configfile.ConfigFile
	EventBus        *eventbus.Bus
	SymbolBlackList SymbolBlackListGetter

	tradeInProgress bool

	coinInfosRefresher *refresher.Refresher[map[string]binance.Symbol]
}

func NewClient(l *log.Logger, cf *configfile.ConfigFile, eb *eventbus.Bus, sbg SymbolBlackListGetter) *Client {
	if cf.TestMode {
		l.Info("Activating Binance test mode")
		binance.UseTestnet = true
	}

	client := Client{
		client:          binance.NewClient(cf.Binance.APIKey, cf.Binance.APIKeySecret),
		Logger:          l,
		ConfigFile:      cf,
		EventBus:        eb,
		SymbolBlackList: sbg,
	}

	client.coinInfosRefresher = refresher.NewRefresher(l, 5*time.Minute, client.RefreshSymbolInfos, refresher.OnErrorLog(client.Logger))

	return &client
}

func (c *Client) LogBalances(ctx context.Context) {
	b, err := c.GetBalance(ctx, append(c.ConfigFile.Coins, c.ConfigFile.Bridge)...)
	if err != nil {
		c.Logger.Error("Failed to get balances", zap.Error(err))
	} else {
		c.Logger.Info(fmt.Sprintf("Balances are %s", util.ToJSON(b)))
	}
}
