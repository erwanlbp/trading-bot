package process

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/erwanlbp/trading-bot/pkg/eventbus"
	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/model"
	"github.com/erwanlbp/trading-bot/pkg/repository.go"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

type PairRatioer struct {
	Logger     *log.Logger
	Repository *repository.Repository
	EventBus   *eventbus.Bus
}

func NewPairRatioer(l *log.Logger, r *repository.Repository, eb *eventbus.Bus) *PairRatioer {
	return &PairRatioer{
		Logger:     l,
		Repository: r,
		EventBus:   eb,
	}
}

func (p *PairRatioer) Start(ctx context.Context) {

	sub := p.EventBus.Subscribe(eventbus.EventCoinsPricesFetched)

	go sub.Handler(p.CalculateRatios)
}

func (p *PairRatioer) CalculateRatios(_ eventbus.Event) {
	logger := p.Logger.With(zap.String("process", "pair_ratioer"))

	lastPrices, err := p.Repository.GetCoinsLastPrice("USDT")
	if err != nil {
		logger.Error("Failed to get coins last price", zap.Error(err))
		return
	}
	if len(lastPrices) == 0 {
		return
	}

	now := lastPrices[0].Timestamp

	pairs, err := p.Repository.GetPairs(repository.ExistingPair())
	if err != nil {
		logger.Error("Failed to get existing pairs", zap.Error(err))
		return
	}

	var pairsHistory []model.PairHistory
	for _, coinFromPrice := range lastPrices {
		for _, coinToPrice := range lastPrices {
			pair, exists := pairs[util.Symbol(coinFromPrice.Coin, coinToPrice.Coin)]
			if !exists {
				continue
			}

			ratio := coinFromPrice.Price / coinToPrice.Price

			history := model.PairHistory{
				PairID:    pair.ID,
				Timestamp: now,
				Ratio:     ratio,
			}
			pairsHistory = append(pairsHistory, history)
		}
	}

	if err := p.Repository.DB.DB.Transaction(func(tx *gorm.DB) error {
		if err := repository.SimpleUpsert(tx, pairsHistory...); err != nil {
			return fmt.Errorf("failed saving pairs history: %w", err)
		}
		return nil
	}); err != nil {
		logger.Error("failed saving ratios", zap.Error(err))
	}

	logger.Debug(fmt.Sprintf("Saved %d pair ratios", len(pairsHistory)))
}
