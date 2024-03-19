package process

import (
	"context"
	"fmt"
	"time"

	"github.com/prprprus/scheduler"
	"go.uber.org/zap"

	"github.com/erwanlbp/trading-bot/pkg/binance"
	"github.com/erwanlbp/trading-bot/pkg/eventbus"
	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/model"
	"github.com/erwanlbp/trading-bot/pkg/repository.go"
)

type PriceGetter struct {
	Logger        *log.Logger
	BinanceClient *binance.Client
	Repository    *repository.Repository
	EventBus      *eventbus.Bus

	AltCoins []string
}

func NewPriceGetter(l *log.Logger, bc *binance.Client, r *repository.Repository, eb *eventbus.Bus, acs []string) *PriceGetter {
	return &PriceGetter{
		Logger:        l,
		BinanceClient: bc,
		Repository:    r,
		EventBus:      eb,
		AltCoins:      acs,
	}
}

func (p *PriceGetter) Start(ctx context.Context) {
	go func() {

		Scheduler, _ := scheduler.NewScheduler(1000)

		id := Scheduler.Every().Second(0).Do(p.FetchCoinsPrices, ctx)

		// To avoid waiting too long before first fetch
		if time.Now().Second() < 20 {
			p.FetchCoinsPrices(ctx)
		}

		// If ctx is canceled, we'll stop the job
		<-ctx.Done()

		if err := Scheduler.CancelJob(id); err != nil {
			p.Logger.Error("failed canceling job", zap.Error(err))
		}
	}()
}

func (p *PriceGetter) FetchCoinsPrices(ctx context.Context) {
	logger := p.Logger.With(zap.String("process", "fetch_price"))
	logger.Debug("Start fetching coins price")

	// TODO Should we get all coins, not just enabled, so that we could run the algorithm on more than just enabled coins to "propose" some new coins
	coins, err := p.Repository.GetEnabledCoins()
	if err != nil {
		logger.Error("Failed to fetch enabled coins, stopping there", zap.Error(err))
		return
	}

	prices, err := p.BinanceClient.GetCoinsPrice(ctx, coins, p.AltCoins)
	if err != nil {
		logger.Error("Failed to get coins prices", zap.Error(err))
		return
	}

	var models []model.CoinPrice
	for _, coinPrice := range prices {
		models = append(models, model.CoinPrice{
			Coin:      coinPrice.Coin,
			AltCoin:   coinPrice.AltCoin,
			Price:     coinPrice.Price,
			Timestamp: coinPrice.Timestamp,
		})
	}

	if err := repository.SimpleUpsert(p.Repository.DB.DB, models...); err != nil {
		logger.Error("Failed to save coin prices", zap.Error(err))
	}

	logger.Debug(fmt.Sprintf("Fetched %d coins price", len(models)))

	p.EventBus.Notify(eventbus.EventCoinsPricesFetched)
}
