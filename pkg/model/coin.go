package model

import "github.com/erwanlbp/trading-bot/pkg/util"

const CoinTableName = "coins"

type Coin struct {
	Coin    string `gorm:"primaryKey"`
	Enabled bool
}

func (Coin) TableName() string {
	return CoinTableName
}

func CoinIDMapper() util.Mapper[Coin, string] {
	return func(c Coin) string {
		return c.Coin
	}
}

func SameCoinPredicate(coin Coin) util.Predicate[Coin] {
	return func(c Coin) bool {
		return c.Coin == coin.Coin
	}
}
