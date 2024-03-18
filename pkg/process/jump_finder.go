package process

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/erwanlbp/trading-bot/pkg/eventbus"
	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/model"
	"github.com/erwanlbp/trading-bot/pkg/repository.go"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

type JumpFinder struct {
	Logger     *log.Logger
	Repository *repository.Repository
	EventBus   *eventbus.Bus
}

func NewJumpFinder(l *log.Logger, r *repository.Repository, eb *eventbus.Bus) *JumpFinder {
	return &JumpFinder{
		Logger:     l,
		Repository: r,
		EventBus:   eb,
	}
}

func (p *JumpFinder) Start(ctx context.Context) {

	sub := p.EventBus.Subscribe(eventbus.EventCoinsPricesFetched)

	go sub.Handler(p.FindJump)
}

func (p *JumpFinder) FindJump(_ eventbus.Event) {
	logger := p.Logger.With(zap.String("process", "jump_finder"))

	// Get pairsRatio from current prices
	pairsRatio, err := p.CalculateRatios()
	if err != nil {
		logger.Error("Failed to calculate new ratios, can't find better coin", zap.Error(err))
		return
	}
	if len(pairsRatio) == 0 {
		logger.Warn("No ratios found (weird), can't find better coin")
		return
	}
	logger.Debug(fmt.Sprintf("Calculated %d pair ratios", len(pairsRatio)))

	currentCoin, err := p.Repository.GetLastJump()
	if err != nil {
		logger.Error("Failed getting current coin", zap.Error(err))
	}

	var pairsFromCurrentCoin []model.PairWithTickerRatio
	for _, p := range pairsRatio {
		if p.Pair.FromCoin == currentCoin.ToCoin {
			pairsFromCurrentCoin = append(pairsFromCurrentCoin, p)
		}
	}

	// TODO Get it from user conf and dynamically !
	threshold := 0.015

	type BJ struct {
		Pair model.PairWithTickerRatio
		Diff float64
	}

	var bestJump *BJ
	for _, pairRatio := range pairsFromCurrentCoin {

		lastPairRatio := pairRatio.Pair.LastJumpOutRatio

		// If we never jumped on this pair, we avg the ratios on the last 15min
		if lastPairRatio == 0 {
			defaultRatio, err := p.Repository.GetAvgLastPairRatioBetween(pairRatio.Pair.ID, pairRatio.Timestamp.Add(-15*time.Minute), pairRatio.Timestamp.Add(-5*time.Second))
			if err != nil {
				logger.Error(fmt.Sprintf("failed to get default ratio for pair %s, ignoring", pairRatio.Pair.LogSymbol()), zap.Error(err))
				continue
			}
			if defaultRatio == 0 {
				logger.Error(fmt.Sprintf("No default ratio found for pair %s, ignoring", pairRatio.Pair.LogSymbol()))
				continue
			}
			lastPairRatio = defaultRatio
		}

		// TODO Add fees

		diff := pairRatio.Ratio / lastPairRatio

		wantedDiff := 1 + threshold

		if diff < wantedDiff {
			logger.Debug(fmt.Sprintf("❌ Pair %s is not good", pairRatio.Pair.LogSymbol()), zap.Float64("current_ratio", pairRatio.Ratio), zap.Float64("last_jump_out_ratio", lastPairRatio), zap.Float64("diff", diff), zap.Float64("threshold", wantedDiff))
			continue
		}

		logger.Debug(fmt.Sprintf("✅ Pair %s is good", pairRatio.Pair.LogSymbol()), zap.Float64("current_ratio", pairRatio.Ratio), zap.Float64("last_jump_out_ratio", lastPairRatio), zap.Float64("diff", diff), zap.Float64("threshold", wantedDiff))

		if bestJump == nil || bestJump.Diff < diff {
			bestJump = &BJ{
				Pair: pairRatio,
				Diff: diff,
			}
		}
	}

	if bestJump == nil {
		logger.Debug(fmt.Sprintf("No jump found from coin %s", currentCoin.ToCoin))
		return
	}

	logger.Info(fmt.Sprintf("Will jump to %s !", bestJump.Pair.Pair.ToCoin))
}

func (p *JumpFinder) CalculateRatios() ([]model.PairWithTickerRatio, error) {
	lastPrices, err := p.Repository.GetCoinsLastPrice("USDT")
	if err != nil {
		return nil, fmt.Errorf("failed to get coins last price: %w", err)
	}
	if len(lastPrices) == 0 {
		return nil, nil
	}

	now := lastPrices[0].Timestamp

	pairs, err := p.Repository.GetPairs(repository.ExistingPair())
	if err != nil {
		return nil, fmt.Errorf("failed to get existing pairs: %w", err)
	}

	var pairsHistory []model.PairHistory
	var res []model.PairWithTickerRatio
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

			res = append(res, model.PairWithTickerRatio{
				Pair:      pair,
				Ratio:     ratio,
				Timestamp: now,
			})
		}
	}

	if err := repository.SimpleUpsert(p.Repository.DB.DB, pairsHistory...); err != nil {
		return nil, fmt.Errorf("failed saving pairs history: %w", err)
	}

	return res, nil
}
