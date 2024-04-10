package handlers

import (
	"context"
	"fmt"

	"gopkg.in/telebot.v3"
)

var (
	notificationMenu = &telebot.ReplyMarkup{ResizeKeyboard: true}
	btnDebug         = notificationMenu.Text("🐞 Debug")
	btnWarn          = notificationMenu.Text("⚠️ Warn")
	btnInfo          = notificationMenu.Text("ℹ️ Info")
	btnError         = notificationMenu.Text("🚨 Error")
	notificationRow  = notificationMenu.Row(btnDebug, btnWarn, btnInfo, btnError)
	notificationRow2 = notificationMenu.Row(btnBackToMainMenu)
)

func (p *Handlers) Notification(ctx context.Context) {
	p.NotificationDebug(ctx)
	p.NotificationWarn(ctx)
	p.NotificationInfo(ctx)
	p.NotificationError(ctx)

	p.TelegramClient.CreateHandler(&btnNotification, func(c telebot.Context) error {
		return c.Send(fmt.Sprintf("Current notification level : %s. Select new one to update it", p.Conf.NotificationLevel), notificationMenu)
	})
}

func (p *Handlers) NotificationDebug(ctx context.Context) {
	p.TelegramClient.CreateHandler(&btnDebug, func(c telebot.Context) error {
		p.Conf.NotificationLevel = "debug"
		return c.Send(fmt.Sprintf("Set notification level to debug"), notificationMenu)
	})
}

func (p *Handlers) NotificationWarn(ctx context.Context) {
	p.TelegramClient.CreateHandler(&btnWarn, func(c telebot.Context) error {
		p.Conf.NotificationLevel = "warn"
		return c.Send(fmt.Sprintf("Set notification level to warn"), notificationMenu)
	})
}

func (p *Handlers) NotificationInfo(ctx context.Context) {
	p.TelegramClient.CreateHandler(&btnInfo, func(c telebot.Context) error {
		p.Conf.NotificationLevel = "info"
		return c.Send(fmt.Sprintf("Set notification level to info"), notificationMenu)
	})
}

func (p *Handlers) NotificationError(ctx context.Context) {
	p.TelegramClient.CreateHandler(&btnError, func(c telebot.Context) error {
		p.Conf.NotificationLevel = "error"
		return c.Send(fmt.Sprintf("Set notification level to error"), notificationMenu)
	})
}
