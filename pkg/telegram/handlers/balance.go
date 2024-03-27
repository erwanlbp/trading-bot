package handlers

import (
	"context"

	"github.com/shopspring/decimal"
	"gopkg.in/telebot.v3"
)

func (p *Handlers) Balance(ctx context.Context) {
	binanceClient := p.BinanceClient
	p.TelegramClient.CreateHandler(&btnBalance, func(c telebot.Context) error {
		balance, err := binanceClient.GetBalance(ctx, append(binanceClient.ConfigFile.Coins, binanceClient.ConfigFile.Bridge)...)
		if err != nil {
			return c.Send("Error while getting balances, please retry")
		}

		balancePositives := ""
		for s, d := range balance {
			if d.GreaterThan(decimal.NewFromFloat(0)) {
				balancePositives = balancePositives + s + ":" + d.String() + ",  "
			}
		}

		err = c.Send("We will display inline button to paginate all balance later")
		if err != nil {
			return c.Send("Error while getting balances, please retry")
		}

		return c.Send(balancePositives)
	})
}
