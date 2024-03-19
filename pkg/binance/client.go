package binance

import (
	"github.com/adshao/go-binance/v2"

	"github.com/erwanlbp/trading-bot/pkg/log"
)

type Client struct {
	client *binance.Client
	Logger *log.Logger
}

func NewClient(l *log.Logger, apiKey, apiSecret string) *Client {
	client := Client{
		client: binance.NewClient(apiKey, apiSecret),
		Logger: l,
	}

	return &client
}
