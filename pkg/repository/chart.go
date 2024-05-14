package repository

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/erwanlbp/trading-bot/pkg/model"
)

func (r *Repository) GetCharts(modifiers ...QueryFilter) ([]model.Chart, error) {

	q := r.DB.DB.Table(model.ChartTableName)

	for _, m := range modifiers {
		q = m(q)
	}

	var res []model.Chart
	err := q.Find(&res).Error

	return res, err
}

func (r *Repository) GetDefaultChartsWithBestDiff() ([]model.Chart, error) {
	cc, _, err := r.GetCurrentCoin()
	if err != nil {
		return nil, fmt.Errorf("failed to get current coint: %w", err)
	}

	bestDiffs, err := r.GetDiff(OrderBy("diff desc"), FromCoin(cc.Coin), Limit(1))
	if err != nil {
		return nil, err
	}
	if len(bestDiffs) == 0 {
		return nil, fmt.Errorf("didn't find any diff")
	}
	bestDiff := bestDiffs[0]
	return []model.Chart{
		{Type: model.ChartTypeCoinPrice, Config: fmt.Sprintf("%s/%s 1", bestDiff.FromCoin, bestDiff.ToCoin)},
		{Type: model.ChartTypeCoinPrice, Config: fmt.Sprintf("%s/%s 3", bestDiff.FromCoin, bestDiff.ToCoin)},
		{Type: model.ChartTypeCoinPrice, Config: fmt.Sprintf("%s/%s 7", bestDiff.FromCoin, bestDiff.ToCoin)},
	}, nil
}

func Type(t string) QueryFilter {
	return func(q *gorm.DB) *gorm.DB {
		return q.Where("type = ?", t)
	}
}

func (r *Repository) SaveNewCoinPriceChart(chartType string, coins []string, days string) (model.Chart, error) {
	var res model.Chart

	if chartType != model.ChartTypeCoinPrice {
		return res, fmt.Errorf("chart type '%s' is not implemented yet", chartType)
	}

	res.Type = model.ChartTypeCoinPrice
	res.Config = fmt.Sprintf("%s %s", strings.Join(coins, ","), days)

	err := SimpleUpsert(r.DB.DB, res)

	return res, err
}
