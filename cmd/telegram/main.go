package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	conf := config.Init()

	logger := conf.Logger

	logger.Debug("Init telegram handlers")
	conf.TelegramHandlers.InitHandlers(ctx)

	logger.Debug("Starting telegram bot")
	conf.TelegramClient.StartBot()

	// Wait until done is closed
	<-done

}
