package main

import (
	"context"
	"time"

	"github.com/erwanlbp/trading-bot/pkg/config"
)

func main() {
	conf := config.Init()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conf.BinanceClient.LogBalances(ctx)
	conf.Repository.LogCurrentCoin()
	conf.BinanceClient.LogUSDTValue(ctx)
}
