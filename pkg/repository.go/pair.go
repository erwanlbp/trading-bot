package repository

import (
	"time"

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

func (r *Repository) GetLastPairRatiosBefore(t time.Time) ([]model.PairHistory, error) {
	var res []model.PairHistory
	err := r.DB.Raw(
		"with cte as (select ph.*, RANK() OVER (partition by pair_id order by `timestamp` desc) as rnk "+
			"from pairs_history ph "+
			"JOIN pairs p on ph.pair_id = p.id "+
			"JOIN coins fc ON p.from_coin = fc.coin "+
			"JOIN coins tc ON p.to_coin = tc.coin "+
			"WHERE ph.timestamp < ? "+
			"AND fc.enabled = 1 AND tc.enabled = 1) "+
			"select pair_id, timestamp, ratio from cte where rnk = 1", t).Debug().Find(&res).Error
	return res, err
}

func (r *Repository) GetAvgLastPairRatioBetween(pairID uint, start, end time.Time) (float64, error) {
	var res float64
	err := r.DB.Select("COALESCE(MIN(ratio), 0)").Table(model.PairHistoryTableName).Where("pair_id = ?", pairID).Where("timestamp BETWEEN ? AND ?", start, end).Find(&res).Error
	return res, err
}

func ExistingPair() QueryFilter {
	return func(q *gorm.DB) *gorm.DB {
		return q.Where("exists = ?", true)
	}
}
