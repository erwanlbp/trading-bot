package model

const BlacklistedSymbolTableName = "blacklisted_symbol"

type BlacklistedSymbol struct {
	Symbol string `gorm:"primaryKey"`
}

func (BlacklistedSymbol) TableName() string {
	return BlacklistedSymbolTableName
}
