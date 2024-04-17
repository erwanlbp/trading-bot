package repository

import (
	"time"

	"github.com/erwanlbp/trading-bot/pkg/model"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

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

func (r *Repository) GetCoinPricesSince(coins []string, altCoin string, from time.Time) ([]model.CoinPrice, error) {
	var data []model.CoinPrice

	req := r.DB.DB.Where("alt_coin = ?", altCoin).Where("timestamp >= ?", from)

	if len(coins) > 1 && !util.Exists(coins, func(coin string) bool { return coin == "*" }) {
		req = req.Where("coin IN ?", coins)
	}

	err := req.Find(&data).Error

	return data, err
}
