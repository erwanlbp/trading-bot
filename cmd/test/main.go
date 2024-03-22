package main

import (
	"context"
	"fmt"

	"github.com/erwanlbp/trading-bot/pkg/config"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

func main() {

	conf := config.Init()

	x, e := conf.BinanceClient.Sell(context.Background(), "IOTA", "USDT")
	fmt.Println(x, e)
	util.DebugPrintJson(x)
}
