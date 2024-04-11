package handlers

import (
	"context"
	"fmt"

	"github.com/erwanlbp/trading-bot/pkg/util"
	"gopkg.in/telebot.v3"
)

var (
	configurationMenu = &telebot.ReplyMarkup{ResizeKeyboard: true}
	btnNotification   = configurationMenu.Text("🔔️ Notification")
	btnReloadConfig   = configurationMenu.Text("♻️ Reload config.yaml")
)

func (p *Handlers) Configuration(ctx context.Context) {
	p.Notification(ctx)
	p.TelegramClient.CreateHandler(&btnReloadConfig, p.ReloadConfigFile)

	p.TelegramClient.CreateHandler(&btnConfiguration, func(c telebot.Context) error {
		return c.Send("What do you want to do ?", mainMenu)
	})
}

func (p *Handlers) ReloadConfigFile(c telebot.Context) error {
	if err := p.GlobalConf.ReloadConfigFile(context.Background()); err != nil {
		return c.Send(fmt.Sprintf("Failed to reload config file: %s", err.Error()))
	}

	response := "Reloaded config file"

	conf := util.Copy(*p.Conf)

	conf.RemoveSecrets()

	confContent := util.ToYAML(conf)
	response += "\n```\n" + confContent + "\n```"

	return c.Send(response, mainMenu, telebot.ModeMarkdown)
}
