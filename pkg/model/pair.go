package model

const PairTableName = "pairs"

type Pair struct {
	ID       uint `gorm:"primaryKey;autoIncrement"`
	FromCoin string
	ToCoin   string
	Ratio    float64
	Exists   bool

	FromCoinDetail Coin `gorm:"foreignKey:FromCoin"`
	ToCoinDetail   Coin `gorm:"foreignKey:ToCoin"`
}

func (Pair) TableName() string {
	return PairTableName
}
