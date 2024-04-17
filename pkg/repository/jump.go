package repository

import (
	"github.com/erwanlbp/trading-bot/pkg/model"
)

func (r *Repository) GetJumps(filters ...QueryFilter) ([]model.Jump, error) {
	var res []model.Jump

	req := r.DB.DB

	for _, f := range filters {
		req = f(req)
	}

	err := req.Find(&res).Error
	if err != nil {
		return res, err
	}

	return res, nil
}
