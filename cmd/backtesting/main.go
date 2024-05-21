package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/erwanlbp/trading-bot/pkg/config"
	"github.com/erwanlbp/trading-bot/pkg/config/globalconf"
	"github.com/erwanlbp/trading-bot/pkg/eventbus"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

func main() {

	// So the user doesn't have to set the variable when starting the exe, but we have it globally
	_ = os.Setenv(globalconf.BACKTESTING_ENV_VAR, "true")

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

	timeStepper := NewTimeStepper(
		// time.Now().Add(-1*30*24*time.Hour).Truncate(time.Minute),
		time.Now().Add(-5*time.Minute).Truncate(time.Minute), // To test
		time.Now().Truncate(time.Minute),
		1*time.Minute,
	)
	util.Now = timeStepper.GetTime

	conf := config.InitBacktesting(ctx)

	logger := conf.Logger

	logger.Debug("Creating the DB if needed")
	if err := conf.DB.MigrateSchema(); err != nil {
		logger.Fatal("failed to migrate DB schema", zap.Error(err))
	}

	logger.Debug("Loading supported coins")
	if err := config.LoadCoins(conf.ConfigFile.Coins, logger, conf.Repository); err != nil {
		logger.Fatal("failed to load supported coins", zap.Error(err))
	}

	logger.Debug("Starting symbol blacklister process")
	conf.ProcessSymbolBlacklister.Start(ctx)

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

	startBalance := 10000

	// Documentation/Hypothesis
	// - We consider the fees are always binance.DefaultFee (0.998001)
	// - We consider the buy/sell orders are instantly executed
	// - We start the simulation with 10000 USDT
	// - We use Binance prod API so that coins have prices even in the past

	// =====
	// Steps
	// =====

	// Before starting, log the config, start/end, etc

	logger.Info("Backtesting simulation; " + timeStepper.String())
	logger.Info(fmt.Sprintf("Coins are %s", conf.ConfigFile.Coins))
	logger.Info(fmt.Sprintf("Starting with %d %v", startBalance, *conf.ConfigFile.StartCoin))

	searchedJumpSub := conf.EventBus.Subscribe(eventbus.EventSearchedJump)
	priceGetter := conf.ProcessPriceGetter

	go func() {
		// Start the backtesting process in 1s by triggering a SearchedJump event
		time.Sleep(1 * time.Second)
		conf.EventBus.Notify(eventbus.SearchedJump())
	}()

	// Wait for the SearchedJump event, trigger the price getter to run one loop
	searchedJumpSub.Handler(ctx, func(ctx context.Context, _ eventbus.Event) {
		timeStepper.Next()

		// Close the sub when we are past the simulation end
		if timeStepper.Done() {
			searchedJumpSub.Close()
			return
		}

		logger.Info(fmt.Sprintf("Current time is %s", timeStepper.GetTime().Format(time.RFC3339)))

		// Price getter process will trigger the jump finder
		priceGetter.FetchCoinsPrices(ctx)
	})

	// After simulation, log the result, nb of jumps, gain %

	// Wait until done is closed
	<-done
}

type TimeStepper struct {
	current time.Time

	start, end time.Time
	step       time.Duration
}

func NewTimeStepper(start, end time.Time, step time.Duration) TimeStepper {
	return TimeStepper{
		start:   start,
		current: start,
		end:     end,
		step:    step,
	}
}

func (t *TimeStepper) GetTime() time.Time {
	return t.current
}

func (t *TimeStepper) Next() {
	t.current = t.current.Add(t.step)
}

func (t *TimeStepper) Done() bool {
	return t.current.After(t.end)
}

func (t *TimeStepper) String() string {
	return fmt.Sprintf("From %s to %s with step %s", t.start, t.end, t.step)
}
