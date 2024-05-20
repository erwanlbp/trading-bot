package process

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/erwanlbp/trading-bot/pkg/binance"
	"github.com/erwanlbp/trading-bot/pkg/config/configfile"
	"github.com/erwanlbp/trading-bot/pkg/eventbus"
	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/model"
	"github.com/erwanlbp/trading-bot/pkg/repository"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

type JumpFinder struct {
	Logger     *log.Logger
	Binance    *binance.Client
	Repository *repository.Repository
	EventBus   *eventbus.Bus
	ConfigFile *configfile.ConfigFile
}

func NewJumpFinder(l *log.Logger,
	r *repository.Repository,
	eb *eventbus.Bus,
	cf *configfile.ConfigFile,
	bc *binance.Client) *JumpFinder {
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
	p.findJump(ctx)

	p.EventBus.Notify(eventbus.GenerateEvent(eventbus.EventSearchedJump, nil))
}

func (p *JumpFinder) findJump(ctx context.Context) {
	logger := p.Logger.With(zap.String("process", "jump_finder"))

	// Get pairsRatio from current prices
	pairsRatio, err := p.CalculateRatios()
	if err != nil {
		logger.Error("Failed to calculate new ratios, can't find better coin", zap.Error(err))
		return
	}

	if p.Binance.IsTradeInProgress() {
		logger.Debug("Stopping jump finder as there is as trade is in progress")
		return
	}

	if len(pairsRatio) == 0 {
		logger.Warn("No ratios found (weird), can't find better coin")
		return
	}

	currentCoin, hasEverJumped, err := p.Repository.GetCurrentCoin()
	if err != nil {
		logger.Error("Failed getting current coin", zap.Error(err))
		return
	}
	// If we never jumped (first init) or something went wrong and we are now back to the bridge
	if !hasEverJumped || currentCoin.Coin == p.ConfigFile.Bridge {
		if !hasEverJumped {
			logger.Info("Never jumped before, will try to find a first coin")
		} else {
			logger.Info("Current coin is the bridge, will try to find a new coin")
		}
		if err := p.FindGoodCoinFromBridge(ctx, pairsRatio); err != nil {
			logger.Error("Failed finding coin from bridge", zap.Error(err))
			return
		}
		p.EventBus.Notify(eventbus.GenerateEvent(eventbus.SaveBalance, nil))
		return
	}

	// Find best pair (if any) to jump

	wantedGain := decimal.NewFromInt(1).Add(p.ConfigFile.Jump.GetNeededGain(currentCoin.Timestamp))

	logger.Debug(fmt.Sprintf("Need a gain of %s", wantedGain))

	type BJ struct {
		Pair model.PairWithTickerRatio
		Diff decimal.Decimal
	}

	var bestJump *BJ
	var computedDiff []model.Diff
	for _, pairRatio := range pairsRatio {

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
			logger.Info(fmt.Sprintf("Pair %s doesn't have a jump ratio yet, defaulting to last 15min avg %s", pairRatio.Pair.LogSymbol(), defaultRatio))
			lastPairRatio = defaultRatio
		}

		feeMultiplier, err := p.Binance.GetJumpFeeMultiplier(ctx, pairRatio.Pair.FromCoin, pairRatio.Pair.ToCoin, p.ConfigFile.Bridge)
		if err != nil {
			feeMultiplier = binance.DefaultFee
		}

		diff := feeMultiplier.Mul(pairRatio.Ratio).Div(lastPairRatio)

		computedDiff = append(computedDiff, model.Diff{
			FromCoin:   pairRatio.Pair.FromCoin,
			ToCoin:     pairRatio.Pair.ToCoin,
			Timestamp:  time.Now().UTC(),
			Diff:       diff,
			NeededDiff: wantedGain,
		})

		if pairRatio.Pair.FromCoin != currentCoin.Coin {
			continue
		}

		if diff.LessThan(wantedGain) {
			logger.Debug(fmt.Sprintf("❌ Pair %s is not good", pairRatio.Pair.LogSymbol()), zap.String("current_ratio", pairRatio.Ratio.String()), zap.String("last_jump_ratio", lastPairRatio.String()), zap.String("diff", diff.String()), zap.String("fee", feeMultiplier.String()), zap.String("threshold", wantedGain.String()))
			continue
		}

		logger.Info(fmt.Sprintf("✅ Pair %s is good", pairRatio.Pair.LogSymbol()), zap.String("current_ratio", pairRatio.Ratio.String()), zap.String("last_jump_ratio", lastPairRatio.String()), zap.String("diff", diff.String()), zap.String("fee", feeMultiplier.String()), zap.String("threshold", wantedGain.String()))

		if bestJump == nil || bestJump.Diff.LessThan(diff) {
			bestJump = &BJ{
				Pair: pairRatio,
				Diff: diff,
			}
		}
	}

	// Clean all data and savec new one to get info about next jump
	err = p.Repository.ReplaceAllDiff(computedDiff)
	if err != nil {
		logger.Warn("Error while updating diff in DB", zap.Error(err))
	}

	if bestJump == nil {
		logger.Debug(fmt.Sprintf("No jump found from coin %s", currentCoin.Coin))
		return
	}

	if err := p.JumpTo(ctx, bestJump.Pair.Pair); err != nil {
		logger.Error("Failed to jump", zap.Error(err))
	}

	p.EventBus.Notify(eventbus.GenerateEvent(eventbus.SaveBalance, nil))
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

	ec, err := p.Repository.GetEnabledCoins()
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled coins: %w", err)
	}
	enabledCoins := util.AsSet(ec, util.Identity[string]())

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

			// We only return pairs which have enabled to_coin, we don't want to jump to some disabled coin
			if enabledCoins[pair.ToCoin] {
				res = append(res, model.PairWithTickerRatio{
					Pair:      pair,
					Ratio:     ratio,
					Timestamp: now,
				})
			}
		}
	}

	if err := repository.SimpleUpsert(p.Repository.DB.DB, pairsHistory...); err != nil {
		return nil, fmt.Errorf("failed saving pairs history: %w", err)
	}

	return res, nil
}

func (p *JumpFinder) JumpTo(ctx context.Context, pair model.Pair) error {
	release, err := p.Binance.TradeLock()
	if err != nil {
		return err
	}
	defer release()

	p.Logger.Info(fmt.Sprintf("Will jump from %s to %s", pair.FromCoin, pair.ToCoin))

	p.Binance.LogBalances(ctx)

	// TODO Add case where the order was created but got timeout with no partial_filled
	// TODO Add case where the order is partially filled but we won't have enough to do next order so we consider it canceled
	sell, err := p.Binance.Sell(ctx, pair.FromCoin, p.ConfigFile.Bridge)
	if err != nil {
		if sell.IsPartiallyExecuted() {
			p.Logger.Warn(fmt.Sprintf("Sell is partially executed, thus we stay on %s and it will be all sold next jump", pair.FromCoin))
			return nil
		}
		p.Logger.Error(fmt.Sprintf("Failed to sell %s", util.LogSymbol(pair.FromCoin, p.ConfigFile.Bridge)), zap.Error(err))
		return err
	}
	// In case something goes wrong afterward, save bridge as current coin
	if _, err := p.Repository.SetCurrentCoin(p.ConfigFile.Bridge, sell.Time()); err != nil {
		p.Logger.Error(fmt.Sprintf("Failed setting current coin to %s during jump, continuing", p.ConfigFile.Bridge), zap.Error(err))
	}
	p.Logger.Info("Sold " + pair.FromCoin)
	// TODO Add case where the order was created but got timeout with no partial_filled
	// TODO Add case where the order is partially filled but we won't have enough to do next order so we consider it canceled
	buy, err := p.Binance.Buy(ctx, pair.ToCoin, p.ConfigFile.Bridge)
	if err != nil {
		if sell.IsPartiallyExecuted() {
			p.Logger.Warn(fmt.Sprintf("Buy is partially executed, thus we go on %s", pair.ToCoin))
		} else {
			p.Logger.Error(fmt.Sprintf("Failed to buy %s", util.LogSymbol(pair.ToCoin, p.ConfigFile.Bridge)), zap.Error(err))
			return err
		}
	}
	p.Logger.Info("Bought " + pair.ToCoin)

	// Save jump and update pairs to new current_coin with new ratio

	jump := model.Jump{
		FromCoin:     pair.FromCoin,
		ToCoin:       pair.ToCoin,
		Timestamp:    buy.Time(),
		FromPrice:    sell.Price(),
		FromQuantity: sell.Quantity(),
		ToPrice:      buy.Price(),
		ToQuantity:   buy.Quantity(),
	}

	if err := repository.SimpleUpsert(p.Repository.DB.DB, jump); err != nil {
		return fmt.Errorf("failed to save jump")
	}
	if err := p.UpdatePairsToCoinRatios(ctx, pair, &buy, &sell); err != nil {
		p.Logger.Error(fmt.Sprintf("Failed to update pairs to coin %s ratios'", pair.ToCoin), zap.Error(err))
		// TODO Not enough
		return err
	}

	return nil
}

// TODO The algo to find a "best" coin could be better lol it's kinda random right now I guess
func (p *JumpFinder) FindGoodCoinFromBridge(ctx context.Context, pairsRatio []model.PairWithTickerRatio) error {
	logger := p.Logger

	// Get last ratios before current tick
	lastRatios, err := p.Repository.GetLastPairRatiosBefore(pairsRatio[0].Timestamp)
	if err != nil {
		logger.Error("Failed to find previous ratios", zap.Error(err))
		return err
	}
	if len(lastRatios) == 0 {
		logger.Warn("Couldn't find previous ratios, will wait next tick to re-check")
		return fmt.Errorf("no ratios")
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

	// TODO The better algo, if I can pull it out of my head
	// 	// Compare last ratios with current ones to find a coin that is going down compared to others
	// 	var bestPair *model.PairWithTickerRatio
	// 	var bestPairLastRatio model.PairHistory
	// 	var bestPairDiff decimal.Decimal

	// 	var doneCoin map[string]bool = make(map[string]bool)

	// 	for _, currentRatio := range pairsRatio {
	// 		if doneCoin[currentRatio.Pair.FromCoin]  {
	// 			continue
	// 		}

	// 		var avg decimal.Decimal

	// 	for _, ratio := range pairsRatio {
	// if ratio.Pair.FromCoin != currentRatio.Pair.FromCoin {
	// 	continue
	// }

	// 		diff := currentRatio.Ratio.Div(lastRatio.Ratio)

	// 		// We want a ratio that is improving
	// 		if diff.LessThan(decimal.NewFromInt(1)) {
	// 			continue
	// 		}

	// 		if diff.GreaterThan(bestPairDiff) {
	// 			bestPair = util.WrapPtr(currentRatio)
	// 			bestPairDiff = diff
	// 			bestPairLastRatio = lastRatio
	// 		}
	// 	}

	if bestPair == nil {
		logger.Info("Couldn't find an interesting coin to buy, skipping this one")
		return fmt.Errorf("nothing interesting")
	}
	release, err := p.Binance.TradeLock()
	if err != nil {
		return err
	}
	defer release()

	bestCoin := bestPair.Pair.FromCoin

	logger.Info(fmt.Sprintf("Best pair from bridge is %s, thus will buy %s", bestPair.Pair.LogSymbol(), bestCoin), zap.String("diff", bestPairDiff.String()), zap.Duration("last_pair_refresh", bestPair.Timestamp.Sub(bestPairLastRatio.Timestamp)))

	buy, err := p.Binance.Buy(ctx, bestCoin, p.ConfigFile.Bridge)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to buy %s", util.LogSymbol(bestCoin, p.ConfigFile.Bridge)), zap.Error(err))
		return err
	}

	if err := p.UpdatePairsToCoinRatios(ctx, model.Pair{ToCoin: bestCoin}, &buy, nil); err != nil {
		logger.Error(fmt.Sprintf("Failed to update pairs to coin %s ratios'", bestCoin), zap.Error(err))
		// TODO Not enough
		return err
	}

	return nil
}

func (p *JumpFinder) UpdatePairsToCoinRatios(ctx context.Context, pair model.Pair, buy, sell *binance.OrderResult) error {

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

	var pairsToSave []model.Pair
	for _, pa := range pairs {
		if pair.FromCoin == pa.FromCoin && pair.ToCoin == pa.ToCoin {
			pa.LastJump = buy.Time()
			pa.LastJumpRatio = sell.Price().Div(buy.Price())
			pa.LastJumpRatioBasedOn = pa.LastJump
		} else {
			pa.LastJumpRatio = prices[util.Symbol(pa.FromCoin, p.ConfigFile.Bridge)].Price.Div(buy.Price())
			pa.LastJumpRatioBasedOn = buy.Time()
		}
		pairsToSave = append(pairsToSave, pa)
	}

	if err := repository.SimpleUpsert(p.Repository.DB.DB, pairsToSave...); err != nil {
		return fmt.Errorf("failed to save pairs ratios: %w", err)
	}

	if _, err := p.Repository.SetCurrentCoin(pair.ToCoin, buy.Time()); err != nil {
		return fmt.Errorf("failed to save current coin: %w", err)
	}

	p.Binance.LogBalances(ctx)

	return nil
}
