package config

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/model"
	"github.com/erwanlbp/trading-bot/pkg/repository.go"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

// Coins for which we'll get the price, to have the value evolution
var AltCoins = []string{"USDT", "BTC"}

func LoadCoins(enabledCoins []string, logger *log.Logger, repo *repository.Repository) error {

	logger.Info(fmt.Sprintf("Found %d supported coins in config file", len(enabledCoins)))

	existingCoins, err := repo.GetAllCoins()
	if err != nil {
		return fmt.Errorf("failed fetching existing coins from DB: %w", err)
	}

	var newAllCoins []model.Coin
	// All coins in the supported coins file are enabled
	for _, coin := range enabledCoins {
		newAllCoins = append(newAllCoins, model.Coin{Coin: coin, Enabled: true})
	}
	// All coins that were previously in DB are now disabled
	for _, existingCoin := range existingCoins {
		if !util.Exists(newAllCoins, model.SameCoinPredicate(existingCoin)) {
			existingCoin.Enabled = false
			newAllCoins = append(newAllCoins, existingCoin)
		}
	}
	if err := repo.DB.Transaction(func(tx *gorm.DB) error {
		if err := repo.DeleteAllCoins(tx); err != nil {
			return fmt.Errorf("failed deleting all coins: %w", err)
		}
		if err := repository.SimpleUpsert(repo, tx, newAllCoins...); err != nil {
			return fmt.Errorf("failed updating coins: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed updating coins: %w", err)
	}

	return nil
}
