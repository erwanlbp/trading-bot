package model

import (
	"time"

	"github.com/shopspring/decimal"
)

const CoinPriceTableName = "coin_price_history"

type CoinPrice struct {
	Coin      string    `gorm:"primaryKey"`
	AltCoin   string    `gorm:"primaryKey"`
	Timestamp time.Time `gorm:"primaryKey"`
	Price     decimal.Decimal
	Averaged  bool

	CoinRef Coin `gorm:"foreignKey:Coin;references:Coin"`
}

func (CoinPrice) TableName() string {
	return CoinPriceTableName
}
