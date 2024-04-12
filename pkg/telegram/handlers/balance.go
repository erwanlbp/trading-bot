package handlers

import (
	"bytes"
	"context"
	"github.com/erwanlbp/trading-bot/pkg/binance"
	"github.com/erwanlbp/trading-bot/pkg/config/configfile"
	"github.com/erwanlbp/trading-bot/pkg/telegram"
	"github.com/erwanlbp/trading-bot/pkg/util"
	"github.com/olekukonko/tablewriter"
	"github.com/shopspring/decimal"
	"gopkg.in/telebot.v3"
	"sort"
)

var (
	USDT = "USDT"
	BTC  = "BTC"
)

func (p *Handlers) Balance(ctx context.Context, conf *configfile.ConfigFile) {
	binanceClient := p.BinanceClient
	p.TelegramClient.CreateHandler(&btnBalance, func(c telebot.Context) error {
		selector := &telebot.ReplyMarkup{}
		balance, err := binanceClient.GetBalance(ctx, append(binanceClient.ConfigFile.Coins, binanceClient.ConfigFile.Bridge)...)
		if err != nil {
			return c.Send("Error while getting balances, please retry")
		}

		balancePositiveCoin := getPositiveBalance(balance)

		altCoinList := []string{USDT, BTC} // TODO put in config file ?
		sort.Strings(altCoinList)
		prices, err := binanceClient.GetCoinsPrice(ctx, balancePositiveCoin, altCoinList)
		if err != nil {
			return c.Send("Error fetching coin price, please retry")
		}

		altValuesByCoin, totalAlt := getAltValueByCoin(prices, balance)

		// Sort to get amount desc but sometimes not really working idk why ?
		keys := util.Keys(altValuesByCoin)
		sort.SliceStable(keys, func(i, j int) bool {
			return altValuesByCoin[keys[i]][0].Price.GreaterThan(altValuesByCoin[keys[j]][0].Price)
		})

		// Build table
		var t = bytes.NewBufferString("")

		chunks := util.Chunk(keys, conf.Telegram.Handlers.NbBalancesDisplayed)
		messagePaginated := map[int]string{}

		headers := append([]string{"Coin", "Value"}, altCoinList...)
		footer := generateFooter(totalAlt)
		for i, chunk := range chunks {
			table := tablewriter.NewWriter(t)
			table.SetBorder(false)
			table.SetHeader(headers)
			table.SetFooter(footer)

			var data [][]string
			for _, coin := range chunk {
				subData := []string{coin}
				coinPrices := altValuesByCoin[coin]
				sort.Slice(coinPrices, func(i, j int) bool {
					return coinPrices[i].AltCoin < coinPrices[j].AltCoin
				})
				subData = append(subData, coin)
				for _, price := range coinPrices {
					subData = append(subData, price.Price.String())
				}
				data = append(data, subData)
			}

			table.AppendBulk(data)
			table.Render()
			messagePaginated[i] = t.String()
			t.Reset()
		}

		buttons := p.CreatePaginatedHandlers(messagePaginated, selector)
		selector.Inline(selector.Row(buttons...))
		return c.Send(telegram.FormatForMD(messagePaginated[0]), selector, telebot.ModeMarkdown)
	})
}

func getAltValueByCoin(prices map[string]binance.CoinPrice, balance map[string]decimal.Decimal) (map[string][]binance.CoinPrice, map[string]decimal.Decimal) {
	altValuesByCoin := map[string][]binance.CoinPrice{}
	totalAlt := map[string]decimal.Decimal{}
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

func getPositiveBalance(balance map[string]decimal.Decimal) []string {
	var balancePositiveCoin []string
	for s, d := range balance {
		if d.GreaterThan(decimal.Zero) {
			balancePositiveCoin = append(balancePositiveCoin, s)
		}
	}
	return balancePositiveCoin
}

func generateFooter(totalAlt map[string]decimal.Decimal) []string {
	totalAltKeys := util.Keys(totalAlt)
	sort.Strings(totalAltKeys)

	var footer = []string{"Total", ""}
	for _, altK := range totalAltKeys {
		if altK == USDT {
			footer = append(footer, totalAlt[altK].Round(2).String())
		} else if altK == BTC {
			footer = append(footer, totalAlt[altK].Round(6).String())
		} else {
			footer = append(footer, totalAlt[altK].Round(4).String())
		}
	}
	return footer
}
