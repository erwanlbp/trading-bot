package main

import (
	"go.uber.org/zap"

	"github.com/erwanlbp/trading-bot/pkg/config"
)

func main() {
	conf := config.Load()

	conf.Logger.Info("Creating the DB if needed")
	if err := conf.DB.MigrateSchema(); err != nil {
		conf.Logger.Fatal("failed to migrate DB schema", zap.Error(err))
	}

	conf.Logger.Info("Loading supported coins")
	if err := config.LoadCoins(conf.Logger, conf.Repository); err != nil {
		conf.Logger.Fatal("failed to load supported coins", zap.Error(err))
	}
}
