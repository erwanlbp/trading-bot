package repository

import (
	"gorm.io/gorm"

	"github.com/erwanlbp/trading-bot/pkg/model"
)

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

func (r *Repository) GetCoinsLastPrice() ([]model.CoinPrice, error) {

	lastTimestampQuery := r.DB.Select("MAX(timestamp)").Table(model.CoinPriceTableName)

	var res []model.CoinPrice
	err := r.DB.Where("timestamp = (?)", lastTimestampQuery).Find(&res).Error
	return res, err
}
