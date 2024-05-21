package process

import (
	"context"

	"github.com/prprprus/scheduler"
	"go.uber.org/zap"

	"github.com/erwanlbp/trading-bot/pkg/binance"
	"github.com/erwanlbp/trading-bot/pkg/constant"
	"github.com/erwanlbp/trading-bot/pkg/eventbus"
	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/model"
	"github.com/erwanlbp/trading-bot/pkg/repository"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

type PriceGetter struct {
	Logger        *log.Logger
	BinanceClient binance.Interface
	Repository    *repository.Repository
	EventBus      *eventbus.Bus

	AltCoins []string
}

func NewPriceGetter(l *log.Logger, bc binance.Interface, r *repository.Repository, eb *eventbus.Bus, acs []string) *PriceGetter {
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

		// If ctx is canceled, we'll stop the job
		<-ctx.Done()

		if err := Scheduler.CancelJob(id); err != nil {
			p.Logger.Error("failed canceling job", zap.Error(err))
		}
	}()
}

func (p *PriceGetter) FetchCoinsPrices(ctx context.Context) {
	logger := p.Logger.With(zap.String("process", "fetch_price"))

	coinModels, err := p.Repository.GetAllCoins()
	if err != nil {
		logger.Error("Failed to fetch enabled coins, stopping there", zap.Error(err))
		return
	}

	var coins []string
	for _, coin := range coinModels {
		coins = append(coins, coin.Coin)
	}

	prices, err := p.BinanceClient.GetCoinsPrice(ctx, coins, []string{constant.USDT})
	if err != nil {
		logger.Error("Failed to get coins prices", zap.Error(err))
		return
	}

	var models []model.CoinPrice
	for _, coinPrice := range prices {
		coin, altCoin, _ := util.Unsymbol(coinPrice.Symbol, coins, []string{constant.USDT})
		models = append(models, model.CoinPrice{
			Coin:      coin,
			AltCoin:   altCoin,
			Price:     coinPrice.Price,
			Timestamp: coinPrice.Timestamp,
		})
	}

	if err := repository.SimpleUpsert(p.Repository.DB.DB, models...); err != nil {
		logger.Error("Failed to save coin prices", zap.Error(err))
	}

	p.EventBus.Notify(eventbus.GenerateEvent(eventbus.EventCoinsPricesFetched, nil))
}
