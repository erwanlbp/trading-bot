package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/erwanlbp/trading-bot/pkg/config"
)

func main() {
	cancelChan := make(chan os.Signal, 1)
	// catch SIGETRM or SIGINTERRUPT
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan bool)

	go func() {
		<-cancelChan
		cancel()
		// Wait 2s to give everyone time to finish
		time.Sleep(2 * time.Second)
		close(done)
	}()

	conf := config.Init(ctx)

	logger := conf.Logger

	logger.Debug("Starting Telegram notification process")
	conf.ProcessTelegramNotifier.Start(ctx)

	logger.Debug("Creating the DB if needed")
	if err := conf.DB.MigrateSchema(); err != nil {
		logger.Fatal("failed to migrate DB schema", zap.Error(err))
	}

	logger.Debug("Loading supported coins")
	if err := config.LoadCoins(conf.ConfigFile.Coins, logger, conf.Repository); err != nil {
		logger.Fatal("failed to load supported coins", zap.Error(err))
	}

	logger.Debug("Loading available pairs")
	if err := conf.Service.InitializePairs(ctx); err != nil {
		logger.Fatal("failed initializing coin pairs", zap.Error(err))
	}

	logger.Debug("Starting telegram client for notification")
	conf.TelegramClient.StartBot()

	logger.Debug("Starting notification process")
	conf.ProcessNotification.Start(ctx)

	logger.Debug("Starting fees getter process")
	conf.ProcessFeeGetter.Start(ctx)

	logger.Debug("Starting jump finder process")
	conf.ProcessJumpFinder.Start(ctx)

	logger.Debug("Starting coins price getter process")
	conf.ProcessPriceGetter.Start(ctx)

	conf.BinanceClient.LogBalances(ctx)
	conf.Repository.LogCurrentCoin()

	// Wait until done is closed
	<-done

}
