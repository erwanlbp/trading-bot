package process

import (
	"context"
	"fmt"

	"github.com/erwanlbp/trading-bot/pkg/eventbus"
	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/model"
	"github.com/erwanlbp/trading-bot/pkg/repository.go"
	"go.uber.org/zap"
)

// Mainly used just to POC the event bus :)

type PriceLogger struct {
	Logger     *log.Logger
	Repository *repository.Repository
	EventBus   *eventbus.Bus
}

func NewPriceLogger(l *log.Logger, r *repository.Repository, eb *eventbus.Bus) *PriceLogger {
	return &PriceLogger{
		Logger:     l,
		Repository: r,
		EventBus:   eb,
	}
}

func (p *PriceLogger) Start(ctx context.Context) {
	go func() {
		logger := p.Logger.With(zap.String("process", "price_logger"))

		sub := p.EventBus.Subscribe(eventbus.EventCoinsPricesFetched)

		for {
			select {
			case <-sub.EventsCh:

				lastPrices, err := p.Repository.GetCoinsLastPrice("")
				if err != nil {
					logger.Error("Failed to get coins last price", zap.Error(err))
					continue
				}

				var groupedByCoin map[string][]model.CoinPrice = make(map[string][]model.CoinPrice)
				for _, coinPrice := range lastPrices {
					groupedByCoin[coinPrice.Coin] = append(groupedByCoin[coinPrice.Coin], coinPrice)
				}

				for coin, coinPrices := range groupedByCoin {
					log := fmt.Sprintf("Coin %s last price: ", coin)
					for _, coinPrice := range coinPrices {
						log = fmt.Sprintf("%s %s:%f", log, coinPrice.AltCoin, coinPrice.Price)
					}
					logger.Debug(log)
				}
			case <-ctx.Done():
				return
			}

		}
	}()
}
