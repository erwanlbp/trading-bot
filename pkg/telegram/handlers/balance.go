package handlers

import (
	"context"
	"sort"
	"time"

	"github.com/shopspring/decimal"
	"gopkg.in/telebot.v3"

	"github.com/erwanlbp/trading-bot/pkg/binance"
	"github.com/erwanlbp/trading-bot/pkg/constant"
	"github.com/erwanlbp/trading-bot/pkg/telegram"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

type balanceDisplayLine struct {
	Coin     string
	Value    decimal.Decimal
	AltValue decimal.Decimal
}

func (p *Handlers) ShowBalances(c telebot.Context) error {
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
	sort.Strings(balancePositiveCoin)

	altCoinList := constant.AltCoins // TODO put in config file ?

	prices, err := p.BinanceClient.GetCoinsPrice(ctx, balancePositiveCoin, altCoinList)
	if err != nil {
		return c.Send("Error fetching coin price, please retry: " + err.Error())
	}

	altValuesByCoin, totalAlt := getAltValueByCoin(prices, balances)

	// Build table
	messagePaginated := map[interface{}]string{}

	altWithCoinPrices := map[string][]balanceDisplayLine{}
	for _, coin := range balancePositiveCoin {
		for _, altPrices := range altValuesByCoin[coin] {
			line := balanceDisplayLine{
				Coin:     coin,
				Value:    balances[coin],
				AltValue: altPrices.Price,
			}
			altWithCoinPrices[altPrices.AltCoin] = append(altWithCoinPrices[altPrices.AltCoin], line)
		}
	}

	p.Logger.Warn(util.ToJSON(altWithCoinPrices))
	p.Logger.Warn(util.ToJSON(totalAlt[constant.BTC]))

	for _, altCoin := range altCoinList {
		headers := []string{"Coin", "Value", altCoin}
		footer := generateFooter(len(headers), altCoin, totalAlt[altCoin])
		p.Logger.Warn(util.ToJSON(footer))
		messagePaginated[altCoin] = util.ToASCIITable(altWithCoinPrices[altCoin], headers, footer, func(line balanceDisplayLine) []string {
			value := line.AltValue.String()
			if altCoin == constant.USDT {
				value = line.AltValue.StringFixed(2)
			}
			return []string{line.Coin, line.Value.String(), value}
		})
	}

	p.Logger.Warn(util.ToJSON(messagePaginated[constant.USDT]))
	p.Logger.Warn(util.ToJSON(messagePaginated[constant.BTC]))

	buttons := p.CreatePaginatedHandlers(messagePaginated, constant.USDT, selector)
	selector.Inline(selector.Row(buttons...))
	return c.Send(telegram.FormatForMD(messagePaginated[constant.USDT]), selector)
}

func getAltValueByCoin(prices map[string]binance.CoinPrice, balance map[string]decimal.Decimal) (map[string][]binance.CoinPrice, map[string]decimal.Decimal) {
	altValuesByCoin := map[string][]binance.CoinPrice{}
	totalAlt := map[string]decimal.Decimal{}

	prices["BTCBTC"] = binance.CoinPrice{
		Coin:    "BTC",
		AltCoin: "BTC",
		Price:   decimal.NewFromInt(1),
	}

	for _, coinPrices := range prices {
		coin := coinPrices.Coin
		balanceValue := balance[coin].Mul(coinPrices.Price)
		cp := binance.CoinPrice{
			Coin:    coin,
			AltCoin: coinPrices.AltCoin,
			Price:   balanceValue,
		}
		altValuesByCoin[coin] = append(altValuesByCoin[coin], cp)
		totalAlt[coinPrices.AltCoin] = totalAlt[coinPrices.AltCoin].Add(balanceValue)
	}
	return altValuesByCoin, totalAlt
}

func generateFooter(headerLen int, altCoin string, total decimal.Decimal) []string {
	var footer = []string{"Total"}

	for i := 0; i < headerLen-2; i++ {
		footer = append(footer, "")
	}

	// Case of very low balance
	if total.LessThan(decimal.NewFromFloat(0.001)) {
		footer = append(footer, "≈0")
	} else if altCoin == constant.USDT {
		footer = append(footer, total.Round(2).String())
	} else {
		footer = append(footer, total.Round(4).String())
	}

	return footer
}
