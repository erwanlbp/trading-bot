package main

import (
	"context"
	"fmt"
	"time"

	"github.com/erwanlbp/trading-bot/pkg/config"
	"github.com/erwanlbp/trading-bot/pkg/util"
	"go.uber.org/zap"
)

func main() {
	conf := config.Init(context.Background())

	if err := conf.DB.MigrateSchema(); err != nil {
		conf.Logger.Fatal("failed to migrate DB schema", zap.Error(err))
	}

	date := time.Date(2024, 3, 30, 12, 2, 12, 0, time.Local)
	conf.Logger.Debug(fmt.Sprintf("fetching price at time %s", date))
	util.DebugPrintJson(conf.BinanceClient.GetSymbolPriceAtTime(context.Background(), "DOGE", "USDT", date))
}
