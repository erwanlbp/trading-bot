package repository

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"github.com/erwanlbp/trading-bot/pkg/db"
)

type Repository struct {
	DB *db.DB
}

func NewRepository(db *db.DB) *Repository {
	return &Repository{
		DB: db,
	}
}

type QueryFilter func(*gorm.DB) *gorm.DB

func SimpleUpsert[T schema.Tabler](tx *gorm.DB, data ...T) error {
	if len(data) == 0 {
		return nil
	}
	return tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(data).Error
}
