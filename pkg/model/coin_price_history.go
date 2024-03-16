package model

import "time"

const CoinPriceTableName = "coin_price_history"

type CoinPrice struct {
	Coin      string    `gorm:"primaryKey"`
	AltCoin   string    `gorm:"primaryKey"`
	Timestamp time.Time `gorm:"primaryKey"`
	Price     float64
}

func (CoinPrice) TableName() string {
	return CoinPriceTableName
}
