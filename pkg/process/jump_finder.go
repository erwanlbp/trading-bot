package process

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/erwanlbp/trading-bot/pkg/binance"
	"github.com/erwanlbp/trading-bot/pkg/config/configfile"
	"github.com/erwanlbp/trading-bot/pkg/eventbus"
	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/model"
	"github.com/erwanlbp/trading-bot/pkg/repository.go"
	"github.com/erwanlbp/trading-bot/pkg/util"
	"github.com/shopspring/decimal"
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

	if p.Binance.IsTradeInProgress() {
		logger.Debug("Stopping there as trade is in progress")
		return
	}

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

	wantedGain := decimal.NewFromInt(1).Add(p.ConfigFile.Jump.GetNeededGain(currentCoin.Timestamp))

	logger.Debug(fmt.Sprintf("Need a gain of %s", wantedGain))

	type BJ struct {
		Pair model.PairWithTickerRatio
		Diff decimal.Decimal
	}

	var bestJump *BJ
	for _, pairRatio := range pairsFromCurrentCoin {

		lastPairRatio := pairRatio.Pair.LastJumpRatio

		// If we never jumped on this pair, we avg the ratios on the last 15min
		// TODO Is it even possible ?
		if pairRatio.Pair.LastJumpRatio.IsZero() {
			defaultRatio, err := p.Repository.GetAvgLastPairRatioBetween(pairRatio.Pair.ID, pairRatio.Timestamp.Add(-15*time.Minute), pairRatio.Timestamp.Add(-5*time.Second))
			if err != nil {
				logger.Error(fmt.Sprintf("failed to get default ratio for pair %s, ignoring", pairRatio.Pair.LogSymbol()), zap.Error(err))
				continue
			}
			if defaultRatio.IsZero() {
				logger.Error(fmt.Sprintf("No default ratio found for pair %s, ignoring", pairRatio.Pair.LogSymbol()))
				continue
			}
			logger.Debug(fmt.Sprintf("Pair %s doesn't have a jump ratio yet, defaulting to last 15min avg %s", pairRatio.Pair.LogSymbol(), defaultRatio))
			lastPairRatio = defaultRatio
		}

		feeMultiplier, err := p.Binance.GetJumpFeeMultiplier(ctx, pairRatio.Pair.FromCoin, pairRatio.Pair.ToCoin, p.ConfigFile.Bridge)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to get jump fee for symbol %s, ignoring", pairRatio.Pair.LogSymbol()), zap.Error(err))
			continue
		}

		diff := feeMultiplier.Mul(pairRatio.Ratio).Div(lastPairRatio)

		if diff.LessThan(wantedGain) {
			// logger.Debug(fmt.Sprintf("❌ Pair %s is not good", pairRatio.Pair.LogSymbol()), zap.Float64("current_ratio", pairRatio.Ratio), zap.Float64("last_jump_out_ratio", lastPairRatio), zap.Float64("diff", diff), zap.Float64("fee", feeMultiplier), zap.Float64("threshold", wantedDiff))
			continue
		}

		logger.Debug(fmt.Sprintf("✅ Pair %s is good", pairRatio.Pair.LogSymbol()), zap.String("current_ratio", pairRatio.Ratio.String()), zap.String("last_jump_ratio", lastPairRatio.String()), zap.String("diff", diff.String()), zap.String("fee", feeMultiplier.String()), zap.String("threshold", wantedGain.String()))

		if bestJump == nil || bestJump.Diff.LessThan(diff) {
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

	if err := p.JumpTo(ctx, bestJump.Pair.Pair); err != nil {
		logger.Error("Failed to jump", zap.Error(err))
	}
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

			ratio := coinFromPrice.Price.Div(coinToPrice.Price)

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

	p.Logger.Debug(fmt.Sprintf("Saved %d pairs ratios", len(pairsHistory)))

	return res, nil
}

func (p *JumpFinder) JumpTo(ctx context.Context, pair model.Pair) error {
	p.Logger.Info(fmt.Sprintf("Will jump from %s to %s", pair.FromCoin, pair.ToCoin))

	sell, err := p.Binance.Sell(ctx, pair.FromCoin, p.ConfigFile.Bridge)
	if err != nil {
		p.Logger.Error(fmt.Sprintf("Failed to sell %s", util.LogSymbol(pair.FromCoin, p.ConfigFile.Bridge)), zap.Error(err))
		return err
	}
	buy, err := p.Binance.Buy(ctx, pair.ToCoin, p.ConfigFile.Bridge)
	if err != nil {
		p.Logger.Error(fmt.Sprintf("Failed to buy %s", util.LogSymbol(pair.ToCoin, p.ConfigFile.Bridge)), zap.Error(err))
		return err
	}

	// Save jump and update pairs to new current_coin with new ratio

	jump := model.Jump{
		FromCoin:  pair.FromCoin,
		ToCoin:    pair.ToCoin,
		Timestamp: time.UnixMilli(buy.Time),
	}

	pairs, err := p.Repository.GetPairs(repository.ToCoin(pair.ToCoin))
	if err != nil {
		return fmt.Errorf("failed to get pairs to new current_coin: %w", err)
	}
	var fromCoins []string
	for _, p := range pairs {
		fromCoins = append(fromCoins, p.FromCoin)
	}

	prices, err := p.Binance.GetCoinsPrice(ctx, fromCoins, []string{p.ConfigFile.Bridge})
	if err != nil {
		return fmt.Errorf("failed to get prices for pair to new current_coin: %w", err)
	}
	var pricesMap map[string]binance.CoinPrice = make(map[string]binance.CoinPrice)
	for _, p := range prices {
		pricesMap[p.Coin] = p
	}

	var pairsToSave []model.Pair
	for _, p := range pairs {
		if p.FromCoin == pair.FromCoin && p.ToCoin == pair.ToCoin {
			p.LastJump = time.UnixMilli(buy.Time)
			p.LastJumpRatio = decimal.RequireFromString(sell.Price).Div(decimal.RequireFromString(buy.Price))
		} else {
			p.LastJumpRatio = pricesMap[p.FromCoin].Price.Div(decimal.RequireFromString(buy.Price))
		}
		pairsToSave = append(pairsToSave, p)
	}

	if err := p.Repository.DB.Transaction(func(tx *gorm.DB) error {
		if err := repository.SimpleUpsert(p.Repository.DB.DB, jump); err != nil {
			return fmt.Errorf("failed to save jump")
		}
		if err := repository.SimpleUpsert(p.Repository.DB.DB, pairsToSave...); err != nil {
			return fmt.Errorf("failed to save pairs ratios")
		}
		return nil
	}); err != nil {
		// TODO If that happen, what do we do ??
		return fmt.Errorf("Failed to update DB after jump: %w", err)
	}

	return nil
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
	var bestPair *model.PairWithTickerRatio
	var bestPairLastRatio model.PairHistory
	var bestPairDiff decimal.Decimal
	for _, currentRatio := range pairsRatio {
		for _, lastRatio := range lastRatios {
			if currentRatio.Pair.ID != lastRatio.PairID {
				continue
			}

			// Ignore the pair if we calculated the ratio too long ago
			if lastRatio.Timestamp.Before(currentRatio.Timestamp.Add(-5 * time.Minute)) {
				continue
			}
			diff := currentRatio.Ratio.Div(lastRatio.Ratio)

			// We want a ratio that is improving
			if diff.LessThan(decimal.NewFromInt(1)) {
				continue
			}

			if diff.GreaterThan(bestPairDiff) {
				bestPair = util.WrapPtr(currentRatio)
				bestPairDiff = diff
				bestPairLastRatio = lastRatio
			}
		}
	}
	if bestPair == nil {
		logger.Info("Couldn't find an interesting coin to buy, skipping this one")
		return nil
	}

	logger.Info(fmt.Sprintf("Best pair for init is %s, thus will buy %s", bestPair.Pair.LogSymbol(), bestPair.Pair.ToCoin), zap.String("diff", bestPairDiff.String()), zap.Duration("last_pair_refresh", bestPair.Timestamp.Sub(bestPairLastRatio.Timestamp)))

	buy, err := p.Binance.Buy(ctx, bestPair.Pair.ToCoin, p.ConfigFile.Bridge)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to buy %s", util.LogSymbol(bestPair.Pair.ToCoin, p.ConfigFile.Bridge)), zap.Error(err))
		return err
	}

	_ = buy

	// TODO Update pairs ratios to new current coin

	return nil
}
