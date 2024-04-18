package repository

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/erwanlbp/trading-bot/pkg/model"
	"github.com/erwanlbp/trading-bot/pkg/util"
	"github.com/shopspring/decimal"
)

func (r *Repository) GetPairs(filters ...QueryFilter) (map[string]model.Pair, error) {
	var pairs []model.Pair

	q := r.DB.DB

	for _, f := range filters {
		q = f(q)
	}

	err := q.Find(&pairs).Error
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
			"select pair_id, timestamp, ratio from cte where rnk = 1", t).Find(&res).Error
	return res, err
}

func (r *Repository) GetAvgLastPairRatioBetween(pairID uint, start, end time.Time) (decimal.Decimal, error) {
	var res decimal.Decimal
	err := r.DB.Select("COALESCE(MIN(ratio), 0)").Table(model.PairHistoryTableName).Where("pair_id = ?", pairID).Where("timestamp BETWEEN ? AND ?", start, end).Find(&res).Error
	return res, err
}

func ExistingPair() QueryFilter {
	return func(q *gorm.DB) *gorm.DB {
		return q.Where("`exists` = 1")
	}
}

func ToCoin(coin string) QueryFilter {
	return func(q *gorm.DB) *gorm.DB {
		return q.Where("to_coin = ?", coin)
	}
}

func (r *Repository) CleanOldPairHistory() (inserted int64, deleted int64, err error) {
	if err := r.DB.Transaction(func(tx *gorm.DB) error {
		resInsert := r.DB.DB.Exec(`
		INSERT OR REPLACE INTO ` + model.PairHistoryTableName + `(pair_id, timestamp, ratio, averaged)
		SELECT pair_id, strftime('%Y-%m-%d %H:00:00',timestamp) AS "timestamp" , avg(ratio) AS ratio , 1 FROM ` + model.PairHistoryTableName + ` ph WHERE (ph.averaged IS NULL OR ph.averaged = 0) AND timestamp < strftime('%Y-%m-%d 00:00:00','now')
		GROUP BY 1,2`)
		if resInsert.Error != nil {
			return fmt.Errorf("failed to insert aggregated data: %w", err)
		}
		inserted = resInsert.RowsAffected

		resDelete := r.DB.DB.Exec(`DELETE FROM ` + model.PairHistoryTableName + ` WHERE (averaged IS NULL OR averaged = 0) AND timestamp < strftime('%Y-%m-%d 00:00:00','now')`)
		if resDelete.Error != nil {
			return fmt.Errorf("failed to delete old data after aggregation: %w", err)
		}
		deleted = resDelete.RowsAffected

		return nil
	}); err != nil {
		return 0, 0, err
	}
	return
}
