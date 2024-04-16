package handlers

import (
	"context"
	"sort"

	"github.com/shopspring/decimal"
	"gopkg.in/telebot.v3"

	"github.com/erwanlbp/trading-bot/pkg/binance"
	"github.com/erwanlbp/trading-bot/pkg/config/configfile"
	_const "github.com/erwanlbp/trading-bot/pkg/const"
	"github.com/erwanlbp/trading-bot/pkg/telegram"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

type balanceDisplayLine struct {
	Coin     string
	Value    decimal.Decimal
	AltValue decimal.Decimal
}

func (p *Handlers) Balance(ctx context.Context, conf *configfile.ConfigFile) {
	binanceClient := p.BinanceClient
	p.TelegramClient.CreateHandler(&btnBalance, func(c telebot.Context) error {
		selector := &telebot.ReplyMarkup{}
		balance, err := binanceClient.GetBalance(ctx, append(binanceClient.ConfigFile.Coins, binanceClient.ConfigFile.Bridge)...)
		if err != nil {
			return c.Send("Error while getting balances, please retry")
		}

		balancePositiveCoin := getPositiveBalance(balance)

		altCoinList := append([]string{}, _const.AltCoins...) // TODO put in config file ?
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
		chunks := util.Chunk(keys, conf.Telegram.Handlers.NbBalancesDisplayed)
		messagePaginated := map[interface{}]string{}

		for _, chunk := range chunks {
			altWithCoinPrices := map[string][]balanceDisplayLine{}
			for _, coin := range chunk {
				for _, altPrices := range altValuesByCoin[coin] {
					line := balanceDisplayLine{
						Coin:     coin,
						Value:    balance[coin],
						AltValue: altPrices.Price,
					}
					altWithCoinPrices[altPrices.AltCoin] = append(altWithCoinPrices[altPrices.AltCoin], line)
				}
			}

			for _, altCoin := range altCoinList {
				headers := []string{"Coin", "Value", altCoin}
				footer := generateFooter(len(headers), altCoin, totalAlt[altCoin])
				messagePaginated[altCoin] = util.ToASCIITable(altWithCoinPrices[altCoin], headers, footer, func(line balanceDisplayLine) []string {
					return []string{line.Coin, line.Value.String(), line.AltValue.String()}
				})
			}
		}

		buttons := p.CreatePaginatedHandlers(messagePaginated, _const.USDT, selector)
		selector.Inline(selector.Row(buttons...))
		return c.Send(telegram.FormatForMD(messagePaginated[_const.USDT]), selector, telebot.ModeMarkdown)
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

func generateFooter(headerLen int, altCoin string, total decimal.Decimal) []string {
	var footer = []string{"Total"}

	for i := 0; i < headerLen-2; i++ {
		footer = append(footer, "")
	}

	// Case of very low balance
	if total.LessThan(decimal.NewFromFloat(0.001)) {
		footer = append(footer, "≈0")
	} else if altCoin == _const.USDT {
		footer = append(footer, total.Round(2).String())
	} else if altCoin == _const.BTC {
		footer = append(footer, total.Round(6).String())
	} else {
		footer = append(footer, total.Round(4).String())
	}

	return footer
}
