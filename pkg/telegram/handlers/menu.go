package handlers

import (
	"context"
	"gopkg.in/telebot.v3"
)

var MenuHandler = telebot.Command{
	Text:        "menu",
	Description: "Display menu",
}

func (p *Handlers) Menu(ctx context.Context) {
	p.TelegramClient.CreateHandler("/menu", func(c telebot.Context) error {
		replyKeyboard := [][]telebot.ReplyButton{
			{
				{
					Text: "⚖️ Balances",
				},
			},
			{
				{
					Text: "🔔 Notifications",
				},
			},
		}

		keyboard := &telebot.ReplyMarkup{
			ReplyKeyboard: replyKeyboard,
		}

		return c.Send("What do you want to do ?", keyboard)
	})
}
