package repository

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/erwanlbp/trading-bot/pkg/model"
)

func (r *Repository) DeleteAllDiff(tx *gorm.DB) error {
	return tx.Exec("DELETE FROM " + model.DiffTableName).Error
}

func (r *Repository) ReplaceAllDiff(computedDiff []model.Diff) error {
	if err := r.DB.Transaction(func(tx *gorm.DB) error {
		if err := r.DeleteAllDiff(tx); err != nil {
			return fmt.Errorf("failed deleting all diff: %w", err)
		}
		if err := SimpleUpsert(tx, computedDiff...); err != nil {
			return fmt.Errorf("failed saving computed diff: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed updating diff: %s", err)
	}
	return nil
}

func (r *Repository) GetDiff(filters ...QueryFilter) ([]model.Diff, error) {
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
