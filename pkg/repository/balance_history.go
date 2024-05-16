package repository

import "github.com/erwanlbp/trading-bot/pkg/model"

func (r *Repository) GetBalanceHistory() ([]model.BalanceHistory, error) {
	var res []model.BalanceHistory
	err := r.DB.Order("timestamp desc").Limit(500).Find(&res).Error
	if err != nil {
		return res, err
	}
	return res, nil
}

func (r *Repository) GetLastBalanceHistory() (model.BalanceHistory, error) {
	var res model.BalanceHistory
	err := r.DB.Order("timestamp desc").Limit(1).Find(&res).Error
	if err != nil {
		return res, err
	}
	return res, nil
}
