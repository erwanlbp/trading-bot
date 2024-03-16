package process

import (
	"context"
	"fmt"

	"github.com/erwanlbp/trading-bot/pkg/eventbus"
	"github.com/erwanlbp/trading-bot/pkg/log"
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

				lastPrices, err := p.Repository.GetCoinsLastPrice()
				if err != nil {
					logger.Error("Failed to get coins last price", zap.Error(err))
					continue
				}

				for _, coinPrice := range lastPrices {
					logger.Debug(fmt.Sprintf("Symbol %s/%s last price: %f", coinPrice.Coin, coinPrice.AltCoin, coinPrice.Price))
				}
			case <-ctx.Done():
				return
			}

		}
	}()
}
