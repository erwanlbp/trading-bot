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

	err := conf.ProcessPriceGetter.Run(context.Background())
	if err != nil {
		logger.Fatal("Failed running price getter process", zap.Error(err))
	}
}
