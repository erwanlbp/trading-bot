package repository

import (
	"fmt"
	"github.com/erwanlbp/trading-bot/pkg/model"
	"gorm.io/gorm"
)

func (r Repository) DeleteAllDiff(tx *gorm.DB) error {
	return tx.Exec("DELETE FROM " + model.DiffTableName).Error
}

func (r Repository) ReplaceAllDiff(computedDiff []model.Diff) {
	if err := r.DB.Transaction(func(tx *gorm.DB) error {
		if err := r.DeleteAllDiff(tx); err != nil {
			return fmt.Errorf("failed deleting all diff: %w", err)
		}
		if err := SimpleUpsert(tx, computedDiff...); err != nil {
			return fmt.Errorf("failed savec computed diff: %w", err)
		}
		return nil
	}); err != nil {
		r.Logger.Error(fmt.Sprintf("failed updating diff: %s", err))
		return
	}
}

func (r Repository) GetDiff(filters ...QueryFilter) ([]model.Diff, error) {
	var diffs []model.Diff

	q := r.DB.DB

	for _, f := range filters {
		q = f(q)
	}

	err := q.Find(&diffs).Error
	if err != nil {
		return nil, err
	}

	return diffs, nil
}
