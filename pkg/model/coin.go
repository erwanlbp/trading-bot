package model

import (
	"time"

	"github.com/erwanlbp/trading-bot/pkg/util"
)

const CoinTableName = "coins"

type Coin struct {
	Coin      string `gorm:"primaryKey"`
	Enabled   bool
	EnabledOn time.Time
}

func (Coin) TableName() string {
	return CoinTableName
}

const CurrentCoinTableName = "current_coin_history"

type CurrentCoin struct {
	Coin      string
	Timestamp time.Time `gorm:"primaryKey"`
}

func (CurrentCoin) TableName() string {
	return CurrentCoinTableName
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
