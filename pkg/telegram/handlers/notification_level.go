package handlers

import (
	"context"
	"gopkg.in/telebot.v3"
)

var NotificationsHandler = telebot.Command{
	Text:        "notifications",
	Description: "Update notification level",
}

func (p *Handlers) NotificationLevel(ctx context.Context, notifLevelConfig []string) {
	p.TelegramClient.CreateHandler("/"+NotificationsHandler.Text, func(c telebot.Context) error {
		return c.Send("ofizehfoizehfofhzh")
	})
}
