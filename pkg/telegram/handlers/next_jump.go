package handlers

import (
	"context"
	"fmt"
	"github.com/erwanlbp/trading-bot/pkg/repository"
	"github.com/erwanlbp/trading-bot/pkg/util"
	"gopkg.in/telebot.v3"
)

func (p *Handlers) NextJump(ctx context.Context) {
	p.TelegramClient.CreateHandler(&btnNextJump, func(c telebot.Context) error {
		pairs, err := p.Repository.GetPairs(repository.ExistingPair())
		if err != nil {
			return c.Send("Error while get next jump info, please retry")
		}
		return c.Send(fmt.Sprintf("Next jump : %s", util.ToJSON(pairs)))
	})
}
