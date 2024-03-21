package repository

import (
	"gorm.io/gorm"

	"github.com/erwanlbp/trading-bot/pkg/model"
)

func (r *Repository) GetLastJump() (model.Jump, bool, error) {
	var res model.Jump
	err := r.DB.Order("timestamp desc").Limit(1).Find(&res).Error
	if err != nil {
		return res, false, err
	}
	if res.ToCoin != "" {
		return res, true, nil
	}

	// Default case, get start coin from config
	if r.ConfigFile.StartCoin != nil {
		return model.Jump{
			ToCoin: *r.ConfigFile.StartCoin,
		}, true, nil
	}

	// If never jumped, leave the coin choice to the algo
	return res, false, nil
}

func (r *Repository) GetAllCoins() ([]model.Coin, error) {
	var res []model.Coin
	err := r.DB.Find(&res).Error
	return res, err
}

func (r *Repository) GetEnabledCoins() ([]string, error) {
	var res []string
	err := r.DB.
		Select("coin").
		Table(model.CoinTableName).
		Where("enabled = ?", true).
		Find(&res).Error
	return res, err
}

func (r *Repository) DeleteAllCoins(tx *gorm.DB) error {
	return tx.Exec("DELETE FROM " + model.CoinTableName).Error
}

func (r *Repository) GetCoinsLastPrice(altCoin string) ([]model.CoinPrice, error) {

	lastTimestampQuery := r.DB.Select("MAX(timestamp)").Table(model.CoinPriceTableName)
	if altCoin != "" {
		lastTimestampQuery = lastTimestampQuery.Where("alt_coin = ?", altCoin)

	}

	var res []model.CoinPrice
	query := r.DB.Where("timestamp = (?)", lastTimestampQuery)
	if altCoin != "" {
		query = query.Where("alt_coin = ?", altCoin)
	}

	err := query.Find(&res).Error
	return res, err
}
