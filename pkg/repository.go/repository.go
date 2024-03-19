package repository

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"github.com/erwanlbp/trading-bot/pkg/config/configfile"
	"github.com/erwanlbp/trading-bot/pkg/db"
)

type Repository struct {
	DB         *db.DB
	ConfigFile *configfile.ConfigFile
}

func NewRepository(db *db.DB, cf *configfile.ConfigFile) *Repository {
	return &Repository{
		DB:         db,
		ConfigFile: cf,
	}
}

type QueryFilter func(*gorm.DB) *gorm.DB

func SimpleUpsert[T schema.Tabler](tx *gorm.DB, data ...T) error {
	if len(data) == 0 {
		return nil
	}
	return tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(data).Error
}
