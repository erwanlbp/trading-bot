package repository

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"github.com/erwanlbp/trading-bot/pkg/config/configfile"
	"github.com/erwanlbp/trading-bot/pkg/db"
	"github.com/erwanlbp/trading-bot/pkg/log"
)

type Repository struct {
	DB         *db.DB
	ConfigFile *configfile.ConfigFile
	Logger     *log.Logger
}

func NewRepository(db *db.DB, cf *configfile.ConfigFile, l *log.Logger) *Repository {
	return &Repository{
		DB:         db,
		ConfigFile: cf,
		Logger:     l,
	}
}

type QueryFilter func(*gorm.DB) *gorm.DB

func SimpleUpsert[T schema.Tabler](tx *gorm.DB, data ...T) error {
	if len(data) == 0 {
		return nil
	}
	return tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(data).Error
}
