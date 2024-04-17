package config

import (
	"fmt"
	"time"

	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/model"
	"github.com/erwanlbp/trading-bot/pkg/repository"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

func LoadCoins(enabledCoins []string, logger *log.Logger, repo *repository.Repository) error {

	logger.Info(fmt.Sprintf("Found %d supported coins in config file", len(enabledCoins)))

	existingCoins, err := repo.GetAllCoins()
	if err != nil {
		return fmt.Errorf("failed fetching existing coins from DB: %w", err)
	}
	existingCoinsMap := util.AsMap(existingCoins, model.CoinIDMapper())

	// All coins in the supported coins file are enabled
	now := time.Now().UTC()
	for _, coin := range enabledCoins {
		existingCoin, alreadyCreated := existingCoinsMap[coin]
		if !alreadyCreated {
			existingCoinsMap[coin] = model.Coin{Coin: coin, Enabled: true, EnabledOn: now}
		} else if !existingCoin.Enabled {
			existingCoin.Enabled = true
			existingCoinsMap[coin] = existingCoin
		}
	}
	// All coins that were previously in DB are now disabled
	for c, coin := range existingCoinsMap {
		if !util.Exists(enabledCoins, func(enabledCoin string) bool { return c == enabledCoin }) {
			coin.Enabled = false
			existingCoinsMap[c] = coin
		}
	}
	if len(existingCoinsMap) == 0 {
		return nil
	}
	if err := repository.SimpleUpsert(repo.DB.DB, util.Values(existingCoinsMap)...); err != nil {
		return fmt.Errorf("failed updating coins: %w", err)
	}

	return nil
}
