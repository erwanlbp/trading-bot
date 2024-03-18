package model

import "time"

const PairTableName = "pairs"
const PairHistoryTableName = "pairs_history"

type Pair struct {
	ID       uint `gorm:"primaryKey;autoIncrement"`
	FromCoin string
	ToCoin   string
	Exists   bool

	FromCoinDetail Coin `gorm:"foreignKey:FromCoin"`
	ToCoinDetail   Coin `gorm:"foreignKey:ToCoin"`
}

func (Pair) TableName() string {
	return PairTableName
}

type PairHistory struct {
	PairID    uint      `gorm:"primaryKey"`
	Timestamp time.Time `gorm:"primaryKey"`
	Ratio     float64

	Pair Pair `gorm:"foreignKey:PairID;references:ID"`
}

func (PairHistory) TableName() string {
	return PairHistoryTableName
}
