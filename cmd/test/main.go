package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/erwanlbp/trading-bot/pkg/binance"
	"github.com/erwanlbp/trading-bot/pkg/config"
	"go.uber.org/zap"
)

func main() {
	cancelChan := make(chan os.Signal, 1)
	// catch SIGETRM or SIGINTERRUPT
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

	ctx := context.Background()

	conf := config.Init(context.Background())

	conf.ProcessTelegramNotifier.Start(ctx)

	logger := conf.Logger

	logger.Debug("this is debug", zap.String("string", "field"))
	logger.Info("this is info", zap.String("string", "field"))
	logger.Info("this is info object", zap.Any("foo", map[string]int{"bar": 2}))
	logger.Info("this is info no fields")
	logger.Error("this is error", zap.String("string", "field"), zap.Error(binance.InvalidFeeValue))
	// logger.Fatal("this is fatal", zap.String("string", "field"))

	<-cancelChan
	time.Sleep(500 * time.Millisecond)
}
