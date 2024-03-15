package binance

import (
	"github.com/adshao/go-binance/v2"
)

type Client struct {
	client *binance.Client
}

func NewClient(apiKey, apiSecret string) *Client {
	client := Client{
		client: binance.NewClient(apiKey, apiSecret),
	}

	return &client
}
