package main

import (
	"context"
	"os"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wcharczuk/go-chart/v2"

	"github.com/erwanlbp/trading-bot/pkg/config"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

func main() {
	conf := config.Init(context.Background())

	coins := []string{"RUNE", "STRK"}
	altCoin := "USDT"

	data, err := conf.Repository.GetCoinPricesSince(coins, altCoin, time.Now().Add(-7*util.Day))
	if err != nil {
		conf.Logger.Fatal(err.Error())
	}

	if len(data) == 0 {
		conf.Logger.Fatal("no price found")
	}

	var minValue map[string]decimal.Decimal = make(map[string]decimal.Decimal)
	var maxValue map[string]decimal.Decimal = make(map[string]decimal.Decimal)

	for _, d := range data {
		symbol := util.LogSymbol(d.Coin, d.AltCoin)
		if min, ok := minValue[symbol]; !ok || d.Price.LessThan(min) {
			minValue[symbol] = d.Price
		}
		if max, ok := maxValue[symbol]; !ok || d.Price.GreaterThan(max) {
			maxValue[symbol] = d.Price
		}
	}

	for i, d := range data {
		symbol := util.LogSymbol(d.Coin, d.AltCoin)
		d.Price = d.Price.Sub(minValue[symbol]).Div(maxValue[symbol].Sub(minValue[symbol])).Mul(decimal.NewFromInt(100))
		data[i] = d
	}

	var series map[string]chart.ContinuousSeries = make(map[string]chart.ContinuousSeries)

	for _, point := range data {
		symbol := util.LogSymbol(point.Coin, point.AltCoin)
		serie, ok := series[symbol]
		if !ok {
			serie = chart.ContinuousSeries{Name: symbol, XValueFormatter: chart.TimeDateValueFormatter, Style: chart.StyleTextDefaults()}
		}
		serie.XValues = append(serie.XValues, float64(point.Timestamp.UnixNano()))
		serie.YValues = append(serie.YValues, point.Price.InexactFloat64())
		series[symbol] = serie
	}

	var finalSeries []chart.Series
	for _, serie := range series {
		finalSeries = append(finalSeries, serie)
	}

	graph := chart.Chart{
		Title:  "foobar",
		Series: finalSeries,
	}
	graph.Elements = []chart.Renderable{chart.LegendThin(&graph)}

	f, err := os.Create("out.png")
	if err != nil {
		conf.Logger.Fatal(err.Error())
	}

	if err := graph.Render(chart.PNG, f); err != nil {
		conf.Logger.Fatal(err.Error())
	}
}
