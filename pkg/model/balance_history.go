package model

import (
	"time"

	"github.com/shopspring/decimal"
)

const BalanceHistoryTableName = "balance_history"

type BalanceHistory struct {
	Timestamp   time.Time `gorm:"primaryKey"`
	BtcBalance  decimal.Decimal
	UsdtBalance decimal.Decimal
}

func (BalanceHistory) TableName() string {
	return BalanceHistoryTableName
}
