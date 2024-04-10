package service

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/erwanlbp/trading-bot/pkg/model"
	"github.com/erwanlbp/trading-bot/pkg/repository"
	"github.com/erwanlbp/trading-bot/pkg/util"
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

	var pairsToSave []model.Pair
	var coinsNeedingPrice []string
	for _, coinFrom := range coins {
		for _, coinTo := range coins {
			if coinFrom == coinTo {
				continue
			}
			pair, ok := allPairs[util.Symbol(coinFrom, coinTo)]
			if ok && !pair.LastJumpRatio.IsZero() {
				continue
			}

			if !ok {
				pair = model.Pair{FromCoin: coinFrom, ToCoin: coinTo, Exists: true}
			}

			coinsNeedingPrice = append(coinsNeedingPrice, pair.FromCoin, pair.ToCoin)

			pairsToSave = append(pairsToSave, pair)
		}
	}

	prices, err := s.Binance.GetCoinsPrice(ctx, coinsNeedingPrice, []string{s.ConfigFile.Bridge})
	if err != nil {
		return fmt.Errorf("failed getting coins prices: %w", err)
	}

	for i, pair := range pairsToSave {
		// TODO  - CORE

		pair.LastJumpRatio = prices[util.Symbol(pair.FromCoin, s.ConfigFile.Bridge)].Price.Div(prices[util.Symbol(pair.ToCoin, s.ConfigFile.Bridge)].Price)
		pairsToSave[i] = pair
	}

	if err := s.Repository.DB.Transaction(func(tx *gorm.DB) error {
		return repository.SimpleUpsert(tx, pairsToSave...)
	}); err != nil {
		return fmt.Errorf("failed saving: %w", err)
	}

	s.Logger.Info(fmt.Sprintf("Initialized %d pairs", len(pairsToSave)))

	return nil
}
