package model

const CoinPriceTableName = "coin_price_history"

type CoinPrice struct {
	Coin string `gorm:"primaryKey"`
}

func (CoinPrice) TableName() string {
	return CoinTableName
}
