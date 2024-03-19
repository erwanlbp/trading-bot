package model

import "time"

const JumpTableName = "jumps"

type Jump struct {
	FromCoin  string    `gorm:"primaryKey"`
	ToCoin    string    `gorm:"primaryKey"`
	Timestamp time.Time `gorm:"primaryKey"`

	FromCoinRef Coin `gorm:"foreignKey:FromCoin;references:Coin"`
	ToCoinRef   Coin `gorm:"foreignKey:ToCoin;references:Coin"`
}

func (Jump) TableName() string {
	return JumpTableName
}
