package handlers

import (
	"context"
	"fmt"

	"gopkg.in/telebot.v3"
)

var (
	notificationMenu = &telebot.ReplyMarkup{ResizeKeyboard: true}
	btnDebug         = notificationMenu.Text("🐞 Debug")
	btnInfo          = notificationMenu.Text("ℹ️ Info")
	btnWarn          = notificationMenu.Text("⚠️ Warn")
	btnError         = notificationMenu.Text("🚨 Error")
)

func (p *Handlers) Notification(c telebot.Context) error {
	return c.Send(fmt.Sprintf("Current notification level : %s. Select new one to update it", p.Conf.NotificationLevel), notificationMenu)
}

func (p *Handlers) NotificationDebug(ctx context.Context) {
	p.TelegramClient.CreateHandler(&btnDebug, func(c telebot.Context) error {
		p.Conf.NotificationLevel = "debug"
		return c.Send("Set notification level to debug", notificationMenu)
	})
}

func (p *Handlers) NotificationWarn(ctx context.Context) {
	p.TelegramClient.CreateHandler(&btnWarn, func(c telebot.Context) error {
		p.Conf.NotificationLevel = "warn"
		return c.Send("Set notification level to warn", notificationMenu)
	})
}

func (p *Handlers) NotificationInfo(ctx context.Context) {
	p.TelegramClient.CreateHandler(&btnInfo, func(c telebot.Context) error {
		p.Conf.NotificationLevel = "info"
		return c.Send("Set notification level to info", notificationMenu)
	})
}

func (p *Handlers) NotificationError(ctx context.Context) {
	p.TelegramClient.CreateHandler(&btnError, func(c telebot.Context) error {
		p.Conf.NotificationLevel = "error"
		return c.Send("Set notification level to error", notificationMenu)
	})
}
