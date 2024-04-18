package main

import (
	"context"
	"fmt"
	"time"

	"github.com/erwanlbp/trading-bot/pkg/config"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

func main() {
	conf := config.Init(context.Background())

	util.DebugPrintJson(conf.BinanceClient.GetSymbolInfos(context.Background(), "DOGEUSDT"))

	date := time.Date(2024, 03, 30, 9, 58, 0, 0, time.UTC)

	for {
		date = date.Add(12 * time.Hour)
		fmt.Println(date)
		util.DebugPrintJson(conf.BinanceClient.GetSymbolPriceAtTime(context.TODO(), "DOGE", "USDT", date))
	}
}
