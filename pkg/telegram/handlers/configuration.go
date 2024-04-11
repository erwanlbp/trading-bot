package handlers

import (
	"context"
	"gopkg.in/telebot.v3"
)

var (
	configurationMenu = &telebot.ReplyMarkup{ResizeKeyboard: true}
	btnNotification   = configurationMenu.Text("🔔️ Notification")
	configurationRow  = configurationMenu.Row(btnNotification, btnBackToMainMenu)
)

func (p *Handlers) Configuration(ctx context.Context) {
	p.Notification(ctx)

	p.TelegramClient.CreateHandler(&btnConfiguration, func(c telebot.Context) error {
		return c.Send("What do you want to do ?", configurationMenu)
	})
}
