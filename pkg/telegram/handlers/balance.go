package handlers

import (
	"context"
	"github.com/erwanlbp/trading-bot/pkg/binance"
	"github.com/shopspring/decimal"
	"gopkg.in/telebot.v3"
)

var BalanceHandler = telebot.Command{
	Text:        "balances",
	Description: "Display balances of selected coin and amount > 0",
}

func (p *Handlers) Balance(ctx context.Context, binanceClient *binance.Client) {
	p.TelegramClient.CreateHandler("/balances", func(c telebot.Context) error {
		balance, err := binanceClient.GetBalance(ctx, append(binanceClient.ConfigFile.Coins, binanceClient.ConfigFile.Bridge)...)
		if err != nil {
			return err
		}

		balancePositives := ""
		for s, d := range balance {
			if d.GreaterThan(decimal.NewFromFloat(0)) {
				balancePositives = balancePositives + s + ":" + d.String() + ",  "
			}
		}

		err = c.Send("Balances present in config file and > 0")
		if err != nil {
			return err
		}

		return c.Send(balancePositives)
	})
}
