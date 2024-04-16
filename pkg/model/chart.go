package model

const ChartTableName = "chart"

const (
	ChartTypeCoinPrice = "coin_price"
)

var DefaultCoinPriceChartAllCoin7Day = Chart{
	Type:   ChartTypeCoinPrice,
	Config: `* 7`,
}

type Chart struct {
	ID     uint `gorm:"primaryKey;autoIncrement"`
	Type   string
	Config string
}

func (Chart) TableName() string {
	return ChartTableName
}
