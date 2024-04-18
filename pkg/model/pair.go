package model

import (
	"time"

	"github.com/erwanlbp/trading-bot/pkg/util"
	"github.com/shopspring/decimal"
)

const PairTableName = "pairs"
const PairHistoryTableName = "pairs_history"

type Pair struct {
	ID       uint `gorm:"primaryKey;autoIncrement"`
	FromCoin string
	ToCoin   string
	Exists   bool

	LastJump      time.Time
	LastJumpRatio decimal.Decimal

	FromCoinDetail Coin `gorm:"foreignKey:FromCoin"`
	ToCoinDetail   Coin `gorm:"foreignKey:ToCoin"`
}

func (Pair) TableName() string {
	return PairTableName
}

func (p Pair) LogSymbol() string {
	return util.LogSymbol(p.FromCoin, p.ToCoin)
}

type PairHistory struct {
	PairID    uint      `gorm:"primaryKey"`
	Timestamp time.Time `gorm:"primaryKey"`
	Ratio     decimal.Decimal
	Averaged  bool

	Pair Pair `gorm:"foreignKey:PairID;references:ID"`
}

func (PairHistory) TableName() string {
	return PairHistoryTableName
}

// TODO Why does this struct is not just PairHistory ?
type PairWithTickerRatio struct {
	Pair      Pair
	Ratio     decimal.Decimal
	Timestamp time.Time
}
