package repository

import (
	"github.com/erwanlbp/trading-bot/pkg/model"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

func (r *Repository) GetAllPairs() (map[string]model.Pair, error) {
	var pairs []model.Pair
	err := r.DB.Find(&pairs).Error
	if err != nil {
		return nil, err
	}

	return util.AsMap(pairs, func(p model.Pair) string { return util.Symbol(p.FromCoin, p.ToCoin) }), nil
}
