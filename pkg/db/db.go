package db

import (
	"gorm.io/gorm"

	"github.com/erwanlbp/trading-bot/pkg/model"
)

type DB struct {
	*gorm.DB
}

func NewDB(db *gorm.DB) *DB {
	return &DB{
		DB: db,
	}
}

func (d *DB) MigrateSchema() error {
	return d.DB.AutoMigrate(
		model.Coin{},
	)
}
