package process

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/erwanlbp/trading-bot/pkg/binance"
	"github.com/erwanlbp/trading-bot/pkg/config/configfile"
	"github.com/erwanlbp/trading-bot/pkg/eventbus"
	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/model"
	"github.com/erwanlbp/trading-bot/pkg/repository.go"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

type JumpFinder struct {
	Logger     *log.Logger
	Binance    *binance.Client
	Repository *repository.Repository
	EventBus   *eventbus.Bus
	ConfigFile *configfile.ConfigFile
}

func NewJumpFinder(l *log.Logger, r *repository.Repository, eb *eventbus.Bus, cf *configfile.ConfigFile, bc *binance.Client) *JumpFinder {
	return &JumpFinder{
		Logger:     l,
		Repository: r,
		EventBus:   eb,
		ConfigFile: cf,
		Binance:    bc,
	}
}

func (p *JumpFinder) Start(ctx context.Context) {

	sub := p.EventBus.Subscribe(eventbus.EventCoinsPricesFetched)

	go sub.Handler(ctx, p.FindJump)
}

func (p *JumpFinder) FindJump(ctx context.Context, _ eventbus.Event) {
	logger := p.Logger.With(zap.String("process", "jump_finder"))

	// Get pairsRatio from current prices
	pairsRatio, err := p.CalculateRatios()
	if err != nil {
		logger.Error("Failed to calculate new ratios, can't find better coin", zap.Error(err))
		return
	}

	// TODO If is "in jump" OR "first init" then stop there

	if len(pairsRatio) == 0 {
		logger.Warn("No ratios found (weird), can't find better coin")
		return
	}

	currentCoin, hasEverJumped, err := p.Repository.GetLastJump()
	if err != nil {
		logger.Error("Failed getting current coin", zap.Error(err))
		return
	}
	if !hasEverJumped {
		logger.Info("Never jumped before, will try to find a first coin")
		if err := p.InitFirstCoinBuy(ctx, pairsRatio); err != nil {
			logger.Error("Failed initializing first coin", zap.Error(err))
			return
		}
		return
	}

	var pairsFromCurrentCoin []model.PairWithTickerRatio
	for _, p := range pairsRatio {
		if p.Pair.FromCoin == currentCoin.ToCoin {
			pairsFromCurrentCoin = append(pairsFromCurrentCoin, p)
		}
	}

	// Find best pair (if any) to jump

	wantedDiff := 1 + p.ConfigFile.Jump.GetNeededGain(currentCoin.Timestamp).InexactFloat64()

	logger.Debug(fmt.Sprintf("Need a ratios change of %f", wantedDiff))

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

		sellingFeePct, err := p.Binance.GetFee(ctx, util.Symbol(pairRatio.Pair.FromCoin, p.ConfigFile.Bridge))
		if err != nil {
			logger.Error(fmt.Sprintf("failed to get selling fee for symbol %s, ignoring", util.LogSymbol(pairRatio.Pair.FromCoin, p.ConfigFile.Bridge)), zap.Error(err))
			continue
		}
		buyingFeePct, err := p.Binance.GetFee(ctx, util.Symbol(pairRatio.Pair.ToCoin, p.ConfigFile.Bridge))
		if err != nil {
			logger.Error(fmt.Sprintf("failed to get buying fee for symbol %s, ignoring", util.LogSymbol(pairRatio.Pair.ToCoin, p.ConfigFile.Bridge)), zap.Error(err))
			continue
		}
		feeMultiplier := 1 - (sellingFeePct + buyingFeePct - (sellingFeePct * buyingFeePct))

		diff := feeMultiplier * pairRatio.Ratio / lastPairRatio

		if diff < wantedDiff {
			// logger.Debug(fmt.Sprintf("❌ Pair %s is not good", pairRatio.Pair.LogSymbol()), zap.Float64("current_ratio", pairRatio.Ratio), zap.Float64("last_jump_out_ratio", lastPairRatio), zap.Float64("diff", diff), zap.Float64("fee", feeMultiplier), zap.Float64("threshold", wantedDiff))
			continue
		}

		logger.Debug(fmt.Sprintf("✅ Pair %s is good", pairRatio.Pair.LogSymbol()), zap.Float64("current_ratio", pairRatio.Ratio), zap.Float64("last_jump_out_ratio", lastPairRatio), zap.Float64("diff", diff), zap.Float64("fee", feeMultiplier), zap.Float64("threshold", wantedDiff))

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

	p.JumpTo(ctx, currentCoin.ToCoin, bestJump.Pair.Pair.ToCoin)
}

func (p *JumpFinder) CalculateRatios() ([]model.PairWithTickerRatio, error) {
	lastPrices, err := p.Repository.GetCoinsLastPrice(p.ConfigFile.Bridge)
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

	p.Logger.Debug(fmt.Sprintf("Updated %d pairs ratios", len(pairsHistory)))

	return res, nil
}

func (p *JumpFinder) JumpTo(ctx context.Context, fromCoin, toCoin string) {
	_, _ = p.Binance.Sell(ctx, fromCoin, p.ConfigFile.Bridge)

	// Sell, to go to bridge
	// - Get balance AVAX & USDT
	// - Get symbol info
	//   - quotePrecision

	// Buy the new coin

}

// TODO The algo to find a "best" coin could be better lol it's kinda random right now I guess
func (p *JumpFinder) InitFirstCoinBuy(ctx context.Context, pairsRatio []model.PairWithTickerRatio) error {
	logger := p.Logger

	// Get last ratios before current tick
	lastRatios, err := p.Repository.GetLastPairRatiosBefore(pairsRatio[0].Timestamp)
	if err != nil {
		logger.Error("Failed to find previous ratios for first jump", zap.Error(err))
		return nil
	}
	if len(lastRatios) == 0 {
		logger.Warn("Couldn't find previous ratios for first jump, will wait next tick to re-check")
		return nil
	}
	// Compare last ratios with current ones to find a coin that is going down compared to others
	var bestPair model.PairWithTickerRatio = pairsRatio[0]
	var bestPairLastRatio model.PairHistory
	var bestPairDiff float64
	for _, currentRatio := range pairsRatio {
		for _, lastRatio := range lastRatios {
			if currentRatio.Pair.ID != lastRatio.PairID {
				continue
			}

			// Ignore the pair if we calculated the ratio too long ago
			if lastRatio.Timestamp.Before(currentRatio.Timestamp.Add(-5 * time.Minute)) {
				continue
			}
			diff := currentRatio.Ratio / lastRatio.Ratio

			// We want a ratio that is improving
			if diff < 1 {
				continue
			}

			if diff > bestPairDiff {
				bestPair = currentRatio
				bestPairDiff = diff
				bestPairLastRatio = lastRatio
			}
		}
	}

	logger.Info(fmt.Sprintf("Best pair for init is %s, thus will buy %s", bestPair.Pair.LogSymbol(), bestPair.Pair.ToCoin), zap.Float64("diff", bestPairDiff), zap.Duration("last_pair_refresh", bestPair.Timestamp.Sub(bestPairLastRatio.Timestamp)))

	res, err := p.Binance.Buy(ctx, bestPair.Pair.ToCoin, p.ConfigFile.Bridge)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to buy %s", util.LogSymbol(bestPair.Pair.ToCoin, p.ConfigFile.Bridge)), zap.Error(err))
		return err
	}
	_ = res

	return nil
}
