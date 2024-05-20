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

	end := time.Now().Truncate(time.Minute)
	start := end.Add(-1 * 30 * 24 * time.Hour)

	// Documentation/Hypothesis
	// - We consider the fees are always binance.DefaultFee (0.998001)
	// - We consider the buy/sell orders are instantly executed
	// - We start the simulation with 10000 USDT
	// - We use Binance prod API so that coins have prices even in the past

	// =====
	// Steps
	// =====

	// Before starting, log the config, start/end, etc

	for current := start; current.Before(end); current = current.Add(1 * time.Minute) {
		// Wait for the SearchedJump event, then directly trigger the price getter
		// Price getter process will trigger the jump finder
	}

	// After simulation, log the result, nb of jumps, gain %

	logger.Debug("Starting Telegram notification process")
	conf.ProcessTelegramNotifier.Start(ctx)

	logger.Debug("Starting symbol blacklister process")
	conf.ProcessSymbolBlacklister.Start(ctx)

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

	logger.Debug("Starting jump finder process")
	conf.ProcessJumpFinder.Start(ctx)

	logger.Debug("Starting coins price getter process")
	conf.ProcessPriceGetter.Start(ctx)

	logger.Debug("Starting cleaner process")
	conf.ProcessCleaner.Start(ctx)

	logger.Debug("Starting save balance process")
	conf.BalanceSaver.Start(ctx)

	// Wait until done is closed
	<-done
}
