package main

import (
	"context"

	"github.com/erwanlbp/trading-bot/pkg/config"
)

func main() {
	conf := config.Init(context.Background())

	conf.ProcessCleaner.CleanPairHistory()
}
