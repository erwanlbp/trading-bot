package service

import (
	"context"
	"fmt"

	"github.com/erwanlbp/trading-bot/pkg/binance"
	"github.com/erwanlbp/trading-bot/pkg/model"
	"github.com/erwanlbp/trading-bot/pkg/repository"
	"github.com/erwanlbp/trading-bot/pkg/util"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// TODO This func seems misplaced, where to put it ?

func (s *Service) InitializePairs(ctx context.Context) error {

	coins, err := s.Repository.GetEnabledCoins()
	if err != nil {
		return fmt.Errorf("failed getting enabled coins: %w", err)
	}

	allPairs, err := s.Repository.GetPairs()
	if err != nil {
		return fmt.Errorf("failed getting existing pairs: %w", err)
	}

	jumps, err := s.Repository.GetJumps()
	if err != nil {
		return fmt.Errorf("failed to get jumps: %w", err)
	}

	lastJumpToCoin := make(map[string]model.Jump)
	for _, jump := range jumps {
		// Safety check
		if jump.Timestamp.IsZero() {
			continue
		}
		if lastJump := lastJumpToCoin[jump.ToCoin]; lastJump.Timestamp.IsZero() || jump.Timestamp.After(lastJump.Timestamp) {
			lastJumpToCoin[jump.ToCoin] = jump
		}
	}

	var pairsNeedingBotStartPriceToSave, pairsNeedingLastJumpPriceToSave []model.Pair
	var coinsNeedingBotStartPrice []string
	for _, coinFrom := range coins {
		for _, coinTo := range coins {
			if coinFrom == coinTo {
				continue
			}
			pair, ok := allPairs[util.Symbol(coinFrom, coinTo)]

			// If pair is in DB and have a ratio, nothing to do
			if ok && !pair.LastJumpRatio.IsZero() {
				continue
			}

			// If pair is in DB but doesn't exists, we don't need to fetch prices
			if ok && !pair.Exists {
				continue
			}

			// If pair is not in DB, we'll create it in DB, and fetch the pair prices to initiate last_jump_ratio
			if !ok {
				pair = model.Pair{FromCoin: coinFrom, ToCoin: coinTo, Exists: true}
			}

			if lastJump, ok := lastJumpToCoin[pair.ToCoin]; ok {
				// If we did already jump to this coin, we initialize the ratio at the last jump timestamp
				// As jumps have different timestamps we have to fetch the prices coin per coin
				fromPrice, err := s.Binance.GetSymbolPriceAtTime(ctx, pair.FromCoin, s.ConfigFile.Bridge, lastJump.Timestamp)
				if err != nil {
					if err == binance.ErrNoPriceFoundAtTime {
						s.Logger.Warn(fmt.Sprintf("Couldn't find price at last jump date for coin %s, disabling it, you'll enable it after next jump, maybe it'll have data", pair.FromCoin))
						if err := s.Repository.DisableCoin(pair.FromCoin); err != nil {
							s.Logger.Error("Failed to disable coin "+pair.FromCoin, zap.Error(err))
							return err
						}
						// Stopping here for this pair, not saving it
						continue
					} else {
						return fmt.Errorf("failed getting from_coin(%s) price at time(%s): %w", pair.FromCoin, lastJump.Timestamp, err)
					}
				}
				toPrice, err := s.Binance.GetSymbolPriceAtTime(ctx, pair.ToCoin, s.ConfigFile.Bridge, lastJump.Timestamp)
				if err != nil {
					if err == binance.ErrNoPriceFoundAtTime {
						s.Logger.Warn(fmt.Sprintf("Couldn't find price at last jump date for coin %s, disabling it, you'll enable it after next jump, maybe it'll have data", pair.FromCoin))
						if err := s.Repository.DisableCoin(pair.FromCoin); err != nil {
							s.Logger.Error("Failed to disable coin "+pair.FromCoin, zap.Error(err))
							return err
						}
						// Stopping here for this pair, not saving it
						continue
					} else {
						return fmt.Errorf("failed getting to_coin(%s) price at time(%s): %w", pair.ToCoin, lastJump.Timestamp, err)
					}
				}

				if !toPrice.Price.Equal(decimal.Zero) {
					pair.LastJumpRatio = fromPrice.Price.Div(toPrice.Price)
					pair.LastJumpRatioBasedOn = lastJump.Timestamp
					pairsNeedingLastJumpPriceToSave = append(pairsNeedingLastJumpPriceToSave, pair)
				}
			} else {
				// If we never jump to this coin, we initialize the ratio at the bot start date
				coinsNeedingBotStartPrice = append(coinsNeedingBotStartPrice, pair.FromCoin, pair.ToCoin)
				pairsNeedingBotStartPriceToSave = append(pairsNeedingBotStartPriceToSave, pair)
			}
		}
	}

	if len(coinsNeedingBotStartPrice) > 0 {
		// Group the price getting for pairs that need "bot start" price
		botStartTime, err := s.Repository.GetBotFirstLaunch()
		if err != nil {
			return fmt.Errorf("failed to get bot first launch datetime: %w", err)
		}
		prices, err := s.Binance.GetCoinsPrice(ctx, coinsNeedingBotStartPrice, []string{s.ConfigFile.Bridge})
		if err != nil {
			return fmt.Errorf("failed getting coins prices: %w", err)
		}
		for i, pair := range pairsNeedingBotStartPriceToSave {
			fromPrice := prices[util.Symbol(pair.FromCoin, s.ConfigFile.Bridge)].Price
			toPrice := prices[util.Symbol(pair.ToCoin, s.ConfigFile.Bridge)].Price
			if !toPrice.Equal(decimal.Zero) {
				pair.LastJumpRatio = fromPrice.Div(toPrice)
				pair.LastJumpRatioBasedOn = botStartTime
				pairsNeedingBotStartPriceToSave[i] = pair
			}
		}
	}

	if err := repository.SimpleUpsert(s.Repository.DB.DB, append(pairsNeedingBotStartPriceToSave, pairsNeedingLastJumpPriceToSave...)...); err != nil {
		return fmt.Errorf("failed saving: %w", err)
	}

	s.Logger.Info(fmt.Sprintf("Initialized %d pairs at 'bot start' ratio, and %d pairs at 'last jump to' ratio", len(pairsNeedingBotStartPriceToSave), len(pairsNeedingLastJumpPriceToSave)))

	return nil
}
