package handlers

import (
	"bytes"
	"strconv"
	"strings"
	"time"

	"gopkg.in/telebot.v3"

	"github.com/shopspring/decimal"
	"github.com/wcharczuk/go-chart/v2"

	"github.com/erwanlbp/trading-bot/pkg/model"
	"github.com/erwanlbp/trading-bot/pkg/repository"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

func (p *Handlers) ChartMenu(c telebot.Context) error {
	charts, err := p.Repository.GetCharts(repository.Type(model.ChartTypeCoinPrice))
	if err != nil {
		return c.Send("Failed to get available charts: " + err.Error())
	}

	if len(charts) == 0 {
		charts = []model.Chart{
			model.DefaultCoinPriceChartAllCoin7Day,
		}
	}

	var btns telebot.Row

	for _, chart := range charts {
		btn := chartMenu.Text(chart.Config)
		p.TelegramClient.CreateHandler(&btn, p.GenerateCoinPriceChart)
		btns = append(btns, btn)
	}
	btns = append(btns,
		btnNewChart,
		btnBackToMainMenu,
	)

	response := &telebot.ReplyMarkup{ResizeKeyboard: true}

	var rows []telebot.Row

	for _, r := range util.Chunk(btns, 3) {
		rows = append(rows, r)
	}

	response.Reply(rows...)

	return c.Send("Saved charts:", response)
}

func (p *Handlers) GenerateCoinPriceChart(c telebot.Context) error {

	parts := strings.Split(c.Text(), " ")
	if len(parts) != 2 {
		return c.Send("Malformed text, should be 'coin1,coin2,coin3 duration_in_day'", chartMenu)
	}

	coins := strings.Split(parts[0], ",")
	days, err := strconv.Atoi(parts[1])
	if err != nil || days <= 0 {
		return c.Send("Malformed duration, should be an int > 0", chartMenu)
	}

	data, err := p.Repository.GetCoinPricesSince(coins, "USDT", time.Now().Add(time.Duration(-1*days)*util.Day))
	if err != nil {
		return c.Send(err.Error())
	}

	if len(data) == 0 {
		return c.Send("no price found")
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
			serie = chart.ContinuousSeries{Name: symbol, XValueFormatter: chart.TimeDateValueFormatter}
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
		Title:  c.Text(),
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
	}

	messageParts = append(messageParts,
		"```",
		"/new_chart type:coin_price coins:COIN1,COIN2,COIN3,COIN4 duration:7",
		"```",
	)

	return c.Send(strings.Join(messageParts, "\n"), telebot.ModeMarkdownV2, telebot.RemoveKeyboard, chartMenu)
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
