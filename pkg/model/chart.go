package model

const ChartTableName = "chart"

const (
	ChartTypeCoinPrice = "coin_price"
)

var DefaultCoinPriceChartAllCoin30Day = Chart{Type: ChartTypeCoinPrice, Config: `* 30`}
var DefaultCoinPriceChartAllCoin7Day = Chart{Type: ChartTypeCoinPrice, Config: `* 7`}
var DefaultCoinPriceChartAllCoin3Day = Chart{Type: ChartTypeCoinPrice, Config: `* 3`}
var DefaultCoinPriceChartAllCoin1Day = Chart{Type: ChartTypeCoinPrice, Config: `* 1`}

type Chart struct {
	ID     uint `gorm:"primaryKey;autoIncrement"`
	Type   string
	Config string
}

func (Chart) TableName() string {
	return ChartTableName
}
