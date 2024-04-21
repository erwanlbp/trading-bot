package model

import (
	"time"

	"github.com/erwanlbp/trading-bot/pkg/util"
	"github.com/shopspring/decimal"
)

const DiffTableName = "diff"

type Diff struct {
	FromCoin   string    `gorm:"primaryKey"`
	ToCoin     string    `gorm:"primaryKey"`
	Timestamp  time.Time `gorm:"primaryKey"`
	Diff       decimal.Decimal
	NeededDiff decimal.Decimal
}

func (Diff) TableName() string {
	return DiffTableName
}

func (d *Diff) LogSymbol() string {
	return util.LogSymbol(d.FromCoin, d.ToCoin)
}
