package handlers

import (
	"bytes"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wcharczuk/go-chart/v2"
	"go.uber.org/zap"
	"gopkg.in/telebot.v3"

	"github.com/erwanlbp/trading-bot/pkg/model"
	"github.com/erwanlbp/trading-bot/pkg/repository"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

func (p *Handlers) ChartMenu(c telebot.Context) error {
	charts, err := p.Repository.GetCharts(repository.Type(model.ChartTypeCoinPrice))
	if err != nil {
		return c.Send("Failed to get available charts: " + err.Error())
	}

	// Default all coins charts
	allCharts := []model.Chart{
		model.DefaultCoinPriceChartAllCoin1Day,
		model.DefaultCoinPriceChartAllCoin7Day,
		model.DefaultCoinPriceChartAllCoin30Day,
	}

	// Default best diff with current coin charts
	bestDiffDefaultCharts, err := p.Repository.GetDefaultChartsWithBestDiff()
	if err != nil {
		p.Logger.Warn("Failed to get best diff for charts, will suggest this graph", zap.Error(err))
	} else {
		allCharts = append(allCharts, bestDiffDefaultCharts...)
	}

	// User charts
	allCharts = append(allCharts, charts...)

	var btns telebot.Row
	for _, chart := range allCharts {
		btn := chartMenu.Text(chart.Config)
		p.TelegramClient.CreateHandler(&btn, p.GenerateChart)
		btns = append(btns, btn)
	}
	btns = append(btns,
		// btnNewChart, // Disabling for now because it's not very useful and well documented
		btnBackToMainMenu,
	)

	response := &telebot.ReplyMarkup{ResizeKeyboard: true, OneTimeKeyboard: true}

	var rows []telebot.Row

	for _, r := range util.Chunk(btns, 3) {
		rows = append(rows, r)
	}

	response.Reply(rows...)

	var message []string = []string{
		"Choose a chart in menu or type a command like ⬇️",
		"`/chart COIN1/COIN2 3`",
		"`/chart COIN1,COIN2,COIN3 1`",
	}

	return c.Send(strings.Join(message, "\n"), response)
}

type ChartPoint struct {
	Serie     string
	Timestamp time.Time
	Value     decimal.Decimal
}

func (p *Handlers) GenerateChart(c telebot.Context) error {

	var parts []string
	if len(c.Args()) == 0 {
		parts = strings.Split(c.Text(), " ")
	} else {
		parts = c.Args()
	}
	if len(parts) != 2 {
		return c.Send("Malformed text, should be 'coin1,coin2,coin3 duration_in_day' or 'coin1/coin2 duration_in_day'", chartMenu)
	}

	days, err := strconv.Atoi(parts[1])
	if err != nil || days <= 0 {
		return c.Send("Malformed duration, should be an int > 0", chartMenu)
	}

	var chartType string
	var data []ChartPoint
	var thresholdLine, jumpThreshold decimal.Decimal

	if strings.Contains(parts[0], "/") {
		chartType = "Pair"
		pairCoins := strings.Split(parts[0], "/")
		symbol := util.LogSymbol(pairCoins[0], pairCoins[1])

		pairs, err := p.Repository.GetPairs(repository.Pair(pairCoins[0], pairCoins[1]))
		if err != nil {
			return c.Send("Failed to get the pair, try again later: "+err.Error(), chartMenu)
		}
		pair := pairs[util.Symbol(pairCoins[0], pairCoins[1])]
		thresholdLine = pair.LastJumpRatio

		diffs, err := p.Repository.GetDiff(repository.FromCoin(pair.FromCoin), repository.ToCoin(pair.ToCoin))
		if err != nil {
			return c.Send("Failed to get the pair diff, try again later: "+err.Error(), chartMenu)
		}
		if len(diffs) > 0 {
			jumpThreshold = diffs[0].NeededDiff.Mul(thresholdLine)
		}

		pairData, err := p.Repository.GetPairRatiosSince(pairCoins[0], pairCoins[1], time.Now().Add(time.Duration(-1*days)*util.Day))
		if err != nil {
			return c.Send(err.Error())
		}
		for _, pd := range pairData {
			data = append(data, ChartPoint{Serie: symbol, Timestamp: pd.Timestamp, Value: pd.Ratio})
		}
	} else {
		chartType = "Coins prices"
		coins := strings.Split(parts[0], ",")

		priceData, err := p.Repository.GetCoinPricesSince(coins, "USDT", time.Now().Add(time.Duration(-1*days)*util.Day))
		if err != nil {
			return c.Send(err.Error())
		}
		for _, pd := range priceData {
			data = append(data, ChartPoint{Serie: util.LogSymbol(pd.Coin, pd.AltCoin), Timestamp: pd.Timestamp, Value: pd.Price})
		}
	}

	if len(data) == 0 {
		return c.Send("no price found")
	}

	var minValue map[string]decimal.Decimal = make(map[string]decimal.Decimal)
	var maxValue map[string]decimal.Decimal = make(map[string]decimal.Decimal)

	for _, d := range data {
		if min, ok := minValue[d.Serie]; !ok || d.Value.LessThan(min) {
			minValue[d.Serie] = d.Value
		}
		if max, ok := maxValue[d.Serie]; !ok || d.Value.GreaterThan(max) {
			maxValue[d.Serie] = d.Value
		}
	}

	// If there are more than 1 serie, we show values as percentage [0,100]
	if len(minValue) > 1 {
		for i, d := range data {
			d.Value = d.Value.Sub(minValue[d.Serie]).Div(maxValue[d.Serie].Sub(minValue[d.Serie])).Mul(decimal.NewFromInt(100))
			data[i] = d
		}
	}

	var series map[string]chart.ContinuousSeries = make(map[string]chart.ContinuousSeries)
	var minX, maxX float64
	for _, point := range data {
		serie, ok := series[point.Serie]
		if !ok {
			serie = chart.ContinuousSeries{Name: point.Serie, XValueFormatter: chart.TimeDateValueFormatter}
		}
		x := float64(point.Timestamp.UnixNano())
		serie.XValues = append(serie.XValues, x)
		serie.YValues = append(serie.YValues, point.Value.InexactFloat64())
		series[point.Serie] = serie

		if minX == 0 || x < minX {
			minX = x
		}
		if maxX == 0 || x > maxX {
			maxX = x
		}
	}

	var finalSeries []chart.Series
	for _, serie := range series {
		finalSeries = append(finalSeries, serie)
	}
	if !thresholdLine.IsZero() {
		// Maybe you'll need to normalize the value if there are multiple series
		// if len(minValue )> 1 {
		// thresholdLine = thresholdLine.Sub(minValue[d.Serie]).Div(maxValue[d.Serie].Sub(minValue[d.Serie])).Mul(decimal.NewFromInt(100))
		// }
		finalSeries = append(finalSeries, &chart.LinearSeries{Name: "Last jump", XValues: []float64{minX, maxX}, InnerSeries: chart.LinearCoefficients(0, thresholdLine.InexactFloat64())})
	}
	if !jumpThreshold.IsZero() {
		finalSeries = append(finalSeries, &chart.LinearSeries{Name: "Next jump", XValues: []float64{minX, maxX}, InnerSeries: chart.LinearCoefficients(0, jumpThreshold.InexactFloat64())})
	}

	graph := chart.Chart{
		Title:  chartType + " " + strings.Join(parts, " "),
		Series: finalSeries,
	}
	graph.Elements = []chart.Renderable{chart.LegendThin(&graph)}

	buffer := bytes.NewBuffer([]byte{})
	if err := graph.Render(chart.PNG, buffer); err != nil {
		return c.Send("Failed generating chart: " + err.Error())
	}

	res := &telebot.Photo{File: telebot.FromReader(buffer)}

	return c.Send(res, mainMenu)
}

func (p *Handlers) NewChart(c telebot.Context) error {
	var messageParts []string = []string{
		"Copy and paste the code",
		"Edit the chart type, coins, duration, and send it to validate",
		"Or ignore this message to do nothing",
		"`/new_chart type:coin_price coins:COIN1,COIN2,COIN3,COIN4 duration:7`",
	}

	return c.Send(strings.Join(messageParts, "\n"), telebot.RemoveKeyboard, chartMenu)
}

func (p *Handlers) ValidateNewChart(c telebot.Context) error {

	var chartType string
	var coins []string
	var duration string

	parts := strings.Split(c.Text(), " ")

	for _, part := range parts {
		splitted := strings.Split(part, ":")
		if len(splitted) != 2 {
			continue
		}
		switch arg := splitted[1]; splitted[0] {
		case "type":
			chartType = arg
		case "coins":
			coins = strings.Split(arg, ",")
		case "duration":
			duration = arg
		default:
			continue
		}
	}

	_, err := p.Repository.SaveNewCoinPriceChart(chartType, coins, duration)
	if err != nil {
		return c.Send("Failed to generate chart: " + err.Error())
	}

	return c.Send("Saved", chartMenu)
}
