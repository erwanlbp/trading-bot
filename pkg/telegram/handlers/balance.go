package handlers

import (
	"bytes"
	"context"
	"sort"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wcharczuk/go-chart/v2"
	"gopkg.in/telebot.v3"

	"github.com/erwanlbp/trading-bot/pkg/constant"
	"github.com/erwanlbp/trading-bot/pkg/telegram"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

func (p *Handlers) ShowBtcBalances(c telebot.Context) error {
	return p.ShowBalances(c, constant.BTC)
}

func (p *Handlers) ShowUsdtBalances(c telebot.Context) error {
	return p.ShowBalances(c, constant.USDT)
}

func (p *Handlers) ShowBalancesWithArg(c telebot.Context) error {
	alt := strings.Split(c.Text(), " ")

	if len(alt) != 2 {
		return c.Send("You can only enter one alt coin")
	}

	return p.ShowBalances(c, strings.ToUpper(alt[1]))
}

func (p *Handlers) ShowBalancesChart(c telebot.Context) error {
	history, err := p.Repository.GetBalanceHistory()
	if err != nil {
		return c.Send("Failed to get balance history")
	}

	if len(history) < 2 {
		return c.Send("Not enough points to generate a chart, try again in few hours")
	}

	var series = make(map[string]chart.ContinuousSeries)

	for _, point := range history {
		x := float64(point.Timestamp.UnixNano())

		sUSDT, ok := series[constant.USDT]
		if !ok {
			sUSDT = chart.ContinuousSeries{Name: constant.USDT, XValueFormatter: chart.TimeDateValueFormatter}
		}
		sBTC, ok := series[constant.BTC]
		if !ok {
			sBTC = chart.ContinuousSeries{Name: constant.BTC, XValueFormatter: chart.TimeDateValueFormatter}
		}

		sUSDT.XValues = append(sUSDT.XValues, x)
		sUSDT.YValues = append(sUSDT.YValues, point.UsdtBalance.InexactFloat64())

		sBTC.XValues = append(sBTC.XValues, x)
		sBTC.YValues = append(sBTC.YValues, point.BtcBalance.InexactFloat64())

		series[constant.USDT] = sUSDT
		series[constant.BTC] = sBTC
	}

	graphUSDT := chart.Chart{
		Title:  "USDT balance history",
		Series: append([]chart.Series{}, series[constant.USDT]),
	}
	graphBTC := chart.Chart{
		Title:  "BTC balance history",
		Series: append([]chart.Series{}, series[constant.BTC]),
	}

	graphUSDT.Elements = []chart.Renderable{chart.LegendThin(&graphUSDT)}
	graphBTC.Elements = []chart.Renderable{chart.LegendThin(&graphBTC)}

	buffer := bytes.NewBuffer([]byte{})
	if err := graphUSDT.Render(chart.PNG, buffer); err != nil {
		return c.Send("Failed generating usdt chart: " + err.Error())
	}
	if err := c.Send(&telebot.Photo{File: telebot.FromReader(buffer)}, balanceMenu); err != nil {
		return c.Send("Failed generating usdt chart: " + err.Error())
	}

	buffer = bytes.NewBuffer([]byte{})
	if err := graphBTC.Render(chart.PNG, buffer); err != nil {
		return c.Send("Failed generating btc chart: " + err.Error())
	}
	res := &telebot.Photo{File: telebot.FromReader(buffer)}
	return c.Send(res, balanceMenu)
}

func (p *Handlers) ShowBalances(c telebot.Context, alt string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	selector := &telebot.ReplyMarkup{}
	balances, err := p.BinanceClient.GetBalance(ctx, append(p.Conf.Coins, p.Conf.Bridge)...)
	if err != nil {
		return c.Send("Error while getting balances, please retry: " + err.Error())
	}

	var balancePositiveCoin []string
	for s, d := range balances {
		if d.GreaterThan(decimal.Zero) {
			balancePositiveCoin = append(balancePositiveCoin, s)
		}
	}

	altCoinList := []string{alt}
	prices, err := p.BinanceClient.GetCoinsPrice(ctx, balancePositiveCoin, altCoinList)
	if err != nil {
		return c.Send("Error fetching coin price, please retry: " + err.Error())
	}

	type balanceDisplayLine struct {
		Coin    string
		Balance decimal.Decimal
		Value   decimal.Decimal
	}
	var balanceValues []balanceDisplayLine
	var totalValue decimal.Decimal

	for coin, balance := range balances {
		if balance.IsZero() {
			continue
		}

		price := prices[util.Symbol(coin, alt)].Price
		if coin == alt {
			price = decimal.NewFromInt(1)
		}

		var line balanceDisplayLine

		line.Coin = coin
		line.Balance = balance
		line.Value = balance.Mul(price)
		balanceValues = append(balanceValues, line)

		totalValue = totalValue.Add(line.Value)
	}

	// Sort by balances value desc
	sort.SliceStable(balanceValues, func(i, j int) bool {
		return balanceValues[i].Value.GreaterThan(balanceValues[j].Value)
	})

	headers := []string{"Coin", "Balance", alt}
	footer := generateFooter(len(headers), alt, totalValue)
	message := util.ToASCIITable(balanceValues, headers, footer, func(line balanceDisplayLine) []string {
		value := line.Value.String()
		if alt == constant.USDT {
			value = line.Value.StringFixed(2)
		}
		return []string{line.Coin, line.Balance.String(), value}
	})

	return c.Send(telegram.FormatForMD(message), selector)
}

func generateFooter(headerLen int, altCoin string, total decimal.Decimal) []string {
	var footer = []string{"Total"}

	for i := 0; i < headerLen-2; i++ {
		footer = append(footer, "")
	}

	// Case of very low balance
	if total.LessThan(decimal.NewFromFloat(0.00001)) {
		footer = append(footer, "≈0")
	} else if altCoin == constant.USDT {
		footer = append(footer, total.Round(2).String())
	} else {
		footer = append(footer, total.Round(4).String())
	}

	return footer
}
