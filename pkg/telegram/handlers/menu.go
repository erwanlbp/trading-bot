package handlers

import (
	"context"

	"gopkg.in/telebot.v3"

	"github.com/erwanlbp/trading-bot/pkg/config/configfile"
)

var (
	btnBackToMainMenu = mainMenu.Text("⬅️ Back")

	// Main menu
	mainMenu         = &telebot.ReplyMarkup{ResizeKeyboard: true}
	btnBalance       = mainMenu.Text("⚖️ Balances")
	btnLast10Jumps   = mainMenu.Text("Last 10 jumps")
	btnNextJump      = mainMenu.Text("⤴️ Next jump")
	btnConfiguration = mainMenu.Text("⚙️ Configuration")

	// Configuration menu
	configurationMenu = &telebot.ReplyMarkup{ResizeKeyboard: true}
	btnListCoins      = configurationMenu.Text("👛 List coins")
	btnEditCoins      = configurationMenu.Text("⛏️ Edit coins")
	btnNotification   = configurationMenu.Text("🔔️ Notification")
	btnReloadConfig   = configurationMenu.Text("♻️ Reload config.yaml")
	btnShowConfigFile = configurationMenu.Text("👀 Show config.yaml")
	btnShowLiveConfig = configurationMenu.Text("🔥 Show live config")
	btnExportDB       = configurationMenu.Text("📬 Export DB")
)

func (p *Handlers) InitMenu(ctx context.Context, conf *configfile.ConfigFile) {
	p.Balance(ctx, conf)
	p.LastTenJumps(ctx)
	p.NextJump(ctx, conf)
	p.Configuration(ctx)
	p.TelegramClient.CreateHandler(&btnBackToMainMenu, p.BackToMainMenu)
	p.TelegramClient.CreateHandler(&btnExportDB, p.ExportDB)

	// Setup menus
	mainMenu.Reply(
		mainMenu.Row(btnBalance, btnLast10Jumps, btnNextJump),
		mainMenu.Row(btnConfiguration),
	)
	configurationMenu.Reply(
		configurationMenu.Row(btnListCoins, btnEditCoins, btnNotification),
		configurationMenu.Row(btnShowLiveConfig, btnShowConfigFile, btnReloadConfig),
		configurationMenu.Row(btnExportDB, btnBackToMainMenu),
	)
	notificationMenu.Reply(
		notificationMenu.Row(btnDebug, btnInfo, btnWarn, btnError),
		notificationMenu.Row(btnBackToMainMenu),
	)

	p.TelegramClient.CreateHandler("/menu", func(c telebot.Context) error {
		return c.Send("What do you want to do ?", mainMenu)
	})
}

func (p *Handlers) BackToMainMenu(c telebot.Context) error {
	return c.Send("Back to menu", mainMenu)
}
