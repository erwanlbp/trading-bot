package handlers

import (
	"context"
	"sort"

	"gopkg.in/telebot.v3"

	"github.com/erwanlbp/trading-bot/pkg/config/configfile"
	"github.com/erwanlbp/trading-bot/pkg/telegram"
	"github.com/erwanlbp/trading-bot/pkg/util"
	"github.com/shopspring/decimal"
)

func (p *Handlers) Balance(ctx context.Context, conf *configfile.ConfigFile) {
	binanceClient := p.BinanceClient
	p.TelegramClient.CreateHandler(&btnBalance, func(c telebot.Context) error {
		selector := &telebot.ReplyMarkup{}
		balance, err := binanceClient.GetBalance(ctx, append(binanceClient.ConfigFile.Coins, binanceClient.ConfigFile.Bridge)...)
		if err != nil {
			return c.Send("Error while getting balances, please retry")
		}

		var balancePositiveCoin []string
		for s, d := range balance {
			if d.GreaterThan(decimal.NewFromFloat(0)) {
				balancePositiveCoin = append(balancePositiveCoin, s)
			}
		}

		altCoinList := []string{"USDT", "BTC"} // TODO put in config file ?
		sort.Strings(altCoinList)
		pricesByAlt, err := binanceClient.GetCoinsPriceGroupByAltCoins(ctx, balancePositiveCoin, altCoinList)
		if err != nil {
			return c.Send("Error fetching coin price, please retry")
		}

		headers := append([]string{"Coin", "Value"}, altCoinList...)
		formatter := telegram.InitFormatter(len(headers))

		keys := util.Keys(pricesByAlt)
		sort.SliceStable(keys, func(i, j int) bool {
			return pricesByAlt[keys[i]].Prices[0].Price.GreaterThan(pricesByAlt[keys[j]].Prices[0].Price)
		})

		chunks := util.Chunk(keys, conf.Telegram.Handlers.NbBalancesDisplayed)

		messagePaginated := map[int]string{}
		for i, chunk := range chunks {
			balanceDisplay := formatter.GenerateHeader(headers)
			balanceDisplay += "\n"
			for _, coin := range chunk {
				prices := pricesByAlt[coin].Prices
				sort.Slice(prices, func(i, j int) bool {
					return prices[i].AltCoin < prices[j].AltCoin
				})
				balanceDisplay += formatter.Resize(coin)
				balanceDisplay += " | " + formatter.Resize(balance[coin].String())
				for _, coinPrice := range prices {
					balanceDisplay += " | " + formatter.Resize(coinPrice.Price.String())
				}
				balanceDisplay += "\n"
			}
			messagePaginated[i] = balanceDisplay
		}

		buttons := p.CreatePaginatedHandlers(messagePaginated, selector)
		selector.Inline(selector.Row(buttons...))
		return c.Send(messagePaginated[0], selector)
	})
}
