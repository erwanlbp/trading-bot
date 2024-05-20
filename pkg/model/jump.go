package model

import (
	"time"

	"github.com/shopspring/decimal"
)

const JumpTableName = "jumps"

type Jump struct {
	FromCoin  string    `gorm:"primaryKey"`
	ToCoin    string    `gorm:"primaryKey"`
	Timestamp time.Time `gorm:"primaryKey"`

	FromQuantity decimal.Decimal
	FromPrice    decimal.Decimal
	ToQuantity   decimal.Decimal
	ToPrice      decimal.Decimal

	FromCoinRef Coin `gorm:"foreignKey:FromCoin;references:Coin"`
	ToCoinRef   Coin `gorm:"foreignKey:ToCoin;references:Coin"`
}

func (Jump) TableName() string {
	return JumpTableName
}
