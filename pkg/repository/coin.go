package repository

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/erwanlbp/trading-bot/pkg/model"
)

func (r *Repository) GetCurrentCoin() (model.CurrentCoin, bool, error) {
	var res model.CurrentCoin
	err := r.DB.Order("timestamp desc").Limit(1).Find(&res).Error
	if err != nil {
		return res, false, err
	}
	if res.Coin != "" {
		return res, true, nil
	}

	// Default case, get start coin from config
	if r.ConfigFile.StartCoin != nil {
		return model.CurrentCoin{
			Coin: *r.ConfigFile.StartCoin,
		}, true, nil
	}

	// If never jumped, leave the coin choice to the algo
	return res, false, nil
}

func (r *Repository) SetCurrentCoin(coin string, ts time.Time) (model.CurrentCoin, error) {
	currentCoin := model.CurrentCoin{
		Coin:      coin,
		Timestamp: ts,
	}
	if err := SimpleUpsert(r.DB.DB, currentCoin); err != nil {
		return currentCoin, fmt.Errorf("failed to save: %w", err)
	}
	return currentCoin, nil
}

func (r *Repository) LogCurrentCoin() {
	cc, hasEverJumped, err := r.GetCurrentCoin()
	if err != nil {
		r.Logger.Error("Failed to get current coin", zap.Error(err))
		return
	}
	if cc.Coin != "" {
		r.Logger.InfoWithNotif("Current coin is " + cc.Coin)
		return
	}

	if !hasEverJumped {
		r.Logger.InfoWithNotif("No current because never jumped")
		return
	}

	r.Logger.Warn("No current but jumped before ? that's weird")
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
