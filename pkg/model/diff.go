package model

import (
	"github.com/shopspring/decimal"
	"time"
)

const DiffTableName = "diff"

type Diff struct {
	FromCoin  string          `gorm:"primaryKey"`
	ToCoin    string          `gorm:"primaryKey"`
	Timestamp time.Time       `gorm:"primaryKey"`
	Diff      decimal.Decimal `gorm:"primaryKey"`
}

func (Diff) TableName() string {
	return DiffTableName
}
