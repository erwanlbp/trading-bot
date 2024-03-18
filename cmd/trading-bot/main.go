package main

import (
	"context"

	"go.uber.org/zap"

	"github.com/erwanlbp/trading-bot/pkg/config"
)

func main() {
	conf := config.Init()

	logger := conf.Logger

	logger.Info("Creating the DB if needed")
	if err := conf.DB.MigrateSchema(); err != nil {
		logger.Fatal("failed to migrate DB schema", zap.Error(err))
	}

	logger.Info("Loading supported coins")
	if err := config.LoadCoins(conf.ConfigFile.Coins, logger, conf.Repository); err != nil {
		logger.Fatal("failed to load supported coins", zap.Error(err))
	}

	logger.Info("Loading available pairs")
	if err := conf.Service.InitializePairs(); err != nil {
		logger.Fatal("failed initializing coin pairs", zap.Error(err))
	}

	// TODO Have a clean exit plan, catch SIGTERM and cancel ctx
	ctx := context.Background()

	logger.Info("Starting jump finder process")
	conf.ProcessJumpFinder.Start(ctx)

	logger.Info("Starting coins price getter process")
	conf.ProcessPriceGetter.Start(ctx)

	<-make(chan int)
}
