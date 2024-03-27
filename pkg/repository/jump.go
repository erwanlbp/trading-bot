package repository

import (
	"github.com/erwanlbp/trading-bot/pkg/model"
)

func (r *Repository) GetJumps(number int) ([]model.Jump, error) {
	var res []model.Jump
	err := r.DB.Order("timestamp desc").Limit(number).Find(&res).Error
	if err != nil {
		return res, err
	}

	return res, nil
}
