package repository

import (
	"time"

	"github.com/erwanlbp/trading-bot/pkg/model"
)

func (r *Repository) GetBotFirstLaunch() (time.Time, error) {
	// We say that the first jump sets the first bot activity
	{
		var res model.Jump
		err := r.DB.DB.Order("timestamp").Limit(1).Find(&res).Error
		if err != nil {
			return time.Time{}, err
		}
		if !res.Timestamp.IsZero() {
			return res.Timestamp, err
		}
	}

	// If we never jumped, we use the first coin_price we have
	{
		var res model.CoinPrice
		err := r.DB.DB.Order("timestamp").Limit(1).Find(&res).Error
		if err != nil {
			return time.Time{}, err
		}
		if !res.Timestamp.IsZero() {
			return res.Timestamp, err
		}
	}

	// If we never jumped nor have coin_price (maybe it is the first launch ?) we'll use Now()
	return time.Now(), nil
}
