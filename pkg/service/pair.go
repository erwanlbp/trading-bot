package service

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/erwanlbp/trading-bot/pkg/model"
	"github.com/erwanlbp/trading-bot/pkg/repository.go"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

func (s *Service) InitializePairs() error {

	coins, err := s.Repository.GetEnabledCoins()
	if err != nil {
		return fmt.Errorf("failed getting enabled coins: %w", err)
	}

	allPairs, err := s.Repository.GetAllPairs()
	if err != nil {
		return fmt.Errorf("failed getting existing pairs: %w", err)
	}

	var newPairs []model.Pair
	for _, coinFrom := range coins {
		for _, coinTo := range coins {
			if coinFrom == coinTo {
				continue
			}
			_, ok := allPairs[util.Symbol(coinFrom, coinTo)]
			if !ok {
				newPairs = append(newPairs, model.Pair{FromCoin: coinFrom, ToCoin: coinTo, Exists: true})
			}
		}
	}

	if err := s.Repository.DB.Transaction(func(tx *gorm.DB) error {
		return repository.SimpleUpsert(s.Repository, tx, newPairs...)
	}); err != nil {
		return fmt.Errorf("failed saving: %w", err)
	}

	s.Logger.Info(fmt.Sprintf("Initialized %d new pairs", len(newPairs)))

	return nil
}
