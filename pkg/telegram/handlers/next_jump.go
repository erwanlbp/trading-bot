package handlers

import (
	"context"
	"fmt"
	"sort"
	"time"

	"gopkg.in/telebot.v3"

	"github.com/erwanlbp/trading-bot/pkg/config/configfile"
	"github.com/erwanlbp/trading-bot/pkg/util"
)

func (p *Handlers) NextJump(ctx context.Context, conf *configfile.ConfigFile) {
	p.TelegramClient.CreateHandler(&btnNextJump, func(c telebot.Context) error {
		selector := &telebot.ReplyMarkup{}
		diff, err := p.Repository.GetDiff()
		if err != nil {
			return c.Send("Error while get next jump info, please retry")
		}

		sort.Slice(diff, func(i, j int) bool {
			return diff[i].Diff.GreaterThan(diff[j].Diff)
		})

		chunks := util.Chunk(diff, conf.Telegram.Handlers.NbDiffDisplayed)

		messagePaginated := map[interface{}]string{}
		for i, chunk := range chunks {
			diffDisplayed := fmt.Sprintf("Diff at : %s \n\n", diff[0].Timestamp.Format(time.DateTime))
			for _, d := range chunk {
				diffDisplayed += d.FromCoin + " ➡️ " + d.ToCoin + " : " + fmt.Sprintf("%.6s", d.Diff) + "\n"
			}
			messagePaginated[i] = diffDisplayed
		}

		buttons := p.CreatePaginatedHandlers(messagePaginated, nil, selector)
		selector.Inline(selector.Row(buttons...))
		return c.Send(messagePaginated[0], selector)
	})
}
