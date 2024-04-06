package handlers

import (
	"context"
	"time"

	"gopkg.in/telebot.v3"
)

func (p *Handlers) LastTenJumps(ctx context.Context) {
	p.TelegramClient.CreateHandler(&btnLast10Jumps, func(c telebot.Context) error {

		jump, err := p.Repository.GetJumps(10)
		if err != nil {
			return c.Send("Error while getting last ten jump, please retry")
		}

		if len(jump) < 1 {
			return c.Send("No jump found in DB")
		}

		resultToPrint := ""
		for _, m := range jump {
			resultToPrint += m.Timestamp.Format(time.RFC3339) + " : " + m.FromCoin + " ➡️ " + m.ToCoin + "\n"
		}

		return c.Send(resultToPrint)
	})
}
