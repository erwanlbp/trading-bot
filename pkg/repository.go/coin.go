package repository

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/erwanlbp/trading-bot/pkg/model"
)

func (r *Repository) GetAllCoins() ([]model.Coin, error) {
	var res []model.Coin
	err := r.DB.Find(&res).Error
	return res, err
}

func (r *Repository) DeleteAllCoins(tx *gorm.DB) error {
	return tx.Exec("DELETE FROM " + model.CoinTableName).Error
}

func (r *Repository) UpsertCoin(tx *gorm.DB, coins ...model.Coin) error {
	if len(coins) == 0 {
		return nil
	}
	return tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(coins).Error
}
