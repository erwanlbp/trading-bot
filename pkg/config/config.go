package config

import (
	"fmt"
	"io"
	"os"

	yaml "gopkg.in/yaml.v3"
	"gorm.io/gorm"

	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/model"
	"github.com/erwanlbp/trading-bot/pkg/repository.go"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

func LoadCoins(logger *log.Logger, repository *repository.Repository) error {

	// Read coins file

	filepath := "config/coins.yaml" // TODO Get it more dynamically ?
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed opening file: %w", err)
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed reading supported coins file: %w", err)
	}

	var data struct {
		Coins []string
	}
	if err := yaml.Unmarshal(content, &data); err != nil {
		return fmt.Errorf("failed unmarshaling coins: %w", err)
	}
	logger.Info(fmt.Sprintf("Found %d supported coins", len(data.Coins)))

	// Update coins in DB

	existingCoins, err := repository.GetAllCoins()
	if err != nil {
		return fmt.Errorf("failed fetching existing coins from DB: %w", err)
	}

	var newAllCoins []model.Coin
	// All coins in the supported coins file are enabled
	for _, coin := range data.Coins {
		newAllCoins = append(newAllCoins, model.Coin{Coin: coin, Enabled: true})
	}
	// All coins that were previously in DB are now disabled
	for _, existingCoin := range existingCoins {
		if !util.Exists(newAllCoins, model.SameCoinPredicate(existingCoin)) {
			existingCoin.Enabled = false
			newAllCoins = append(newAllCoins, existingCoin)
		}
	}
	if err := repository.DB.Transaction(func(tx *gorm.DB) error {
		if err := repository.DeleteAllCoins(tx); err != nil {
			return fmt.Errorf("failed deleting all coins: %w", err)
		}
		if err := repository.UpsertCoin(tx, newAllCoins...); err != nil {
			return fmt.Errorf("failed updating coins: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed updating coins: %w", err)
	}

	return nil
}
