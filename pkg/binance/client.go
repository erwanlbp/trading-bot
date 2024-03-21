package binance

import (
	"time"

	"github.com/adshao/go-binance/v2"

	"github.com/erwanlbp/trading-bot/pkg/config/configfile"
	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/refresher"
)

type Client struct {
	client     *binance.Client
	Logger     *log.Logger
	ConfigFile *configfile.ConfigFile

	tradeInProgress bool

	coinInfosRefresher *refresher.Refresher[map[string]binance.Symbol]
}

func NewClient(l *log.Logger, cf *configfile.ConfigFile, apiKey, apiSecret string) *Client {
	client := Client{
		client:     binance.NewClient(apiKey, apiSecret),
		Logger:     l,
		ConfigFile: cf,
	}

	client.coinInfosRefresher = refresher.NewRefresher(l, 5*time.Minute, client.RefreshSymbolInfos, refresher.OnErrorLog(client.Logger))

	return &client
}
