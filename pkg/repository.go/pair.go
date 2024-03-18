package repository

import (
	"gorm.io/gorm"

	"github.com/erwanlbp/trading-bot/pkg/model"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

func (r *Repository) GetPairs(filters ...QueryFilter) (map[string]model.Pair, error) {
	var pairs []model.Pair

	q := r.DB.DB

	for _, f := range filters {
		q = f(q)
	}

	err := r.DB.Find(&pairs).Error
	if err != nil {
		return nil, err
	}

	return util.AsMap(pairs, func(p model.Pair) string { return util.Symbol(p.FromCoin, p.ToCoin) }), nil
}

func ExistingPair() QueryFilter {
	return func(q *gorm.DB) *gorm.DB {
		return q.Where("exists = ?", true)
	}
}
