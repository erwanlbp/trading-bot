package repository

import (
	"gorm.io/gorm/clause"

	"github.com/erwanlbp/trading-bot/pkg/model"
)

func (r *Repository) BlacklistSymbol(symbol string) error {
	return r.DB.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(model.BlacklistedSymbol{Symbol: symbol}).Error
}

func (r *Repository) GetBlacklistedSymbols() ([]model.BlacklistedSymbol, error) {
	var res []model.BlacklistedSymbol
	err := r.DB.DB.Find(&res).Error
	return res, err
}
