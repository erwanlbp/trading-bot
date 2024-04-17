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
