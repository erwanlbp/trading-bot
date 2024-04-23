package handlers

import (
	"context"
	"strings"

	"gopkg.in/telebot.v3"
)

var (
	btnBackToMainMenu = mainMenu.Text("⬅️ Back")

	// Main menu
	mainMenu         = &telebot.ReplyMarkup{ResizeKeyboard: true}
	btnBalance       = mainMenu.Text("⚖️ Balances")
	btnLast10Jumps   = mainMenu.Text("🦘 Last jumps")
	btnNextJump      = mainMenu.Text("⤴️ Next jump")
	btnConfiguration = mainMenu.Text("⚙️ Configuration")
	btnChart         = mainMenu.Text("📊 Chart")

	// Chart menu
	chartMenu   = &telebot.ReplyMarkup{ResizeKeyboard: true}
	btnNewChart = mainMenu.Text("➕ New")

	// Configuration menu
	configurationMenu = &telebot.ReplyMarkup{ResizeKeyboard: true}
	btnListCoins      = configurationMenu.Text("👛 List coins")
	btnEditCoins      = configurationMenu.Text("⛏️ Edit coins")
	btnEditJump       = configurationMenu.Text("🦘 Edit Jump")
	btnNotification   = configurationMenu.Text("🔔️ Notification")
	btnReloadConfig   = configurationMenu.Text("♻️ Reload config.yaml")
	btnShowConfigFile = configurationMenu.Text("👀 Show config.yaml")
	btnShowLiveConfig = configurationMenu.Text("🔥 Show live config")
	btnExportDB       = configurationMenu.Text("📬 Export DB")
)

var availableCommands = []string{
	"/help",
	"/balances",
	"/last_jumps",
	"/next_jump",
	"/new_chart",
	"/chart COIN1/COIN2 3",
	"/chart COIN1,COIN2,COIN3 3",
	"/export_db",
	"/reload_config",
	"/live_config",
	"/config_file",
	"/list_coins",
	"/edit_coins COIN1,COIN2,COIN3",
	"/edit_jump when:3 decrease:0.1 after:1h min:0.1",
}

func (p *Handlers) InitMenu(ctx context.Context) {
	p.TelegramClient.CreateHandler("/help", p.Help)

	p.TelegramClient.CreateHandler(&btnBackToMainMenu, p.BackToMainMenu)

	p.TelegramClient.CreateHandler("/balances", p.ShowBalances)
	p.TelegramClient.CreateHandler(&btnBalance, p.ShowBalances)
	p.TelegramClient.CreateHandler("/last_jumps", p.LastTenJumps)
	p.TelegramClient.CreateHandler(&btnLast10Jumps, p.LastTenJumps)
	p.TelegramClient.CreateHandler("/next_jump", p.NextJump)
	p.TelegramClient.CreateHandler(&btnNextJump, p.NextJump)

	p.TelegramClient.CreateHandler(&btnChart, p.ChartMenu)
	p.TelegramClient.CreateHandler(&btnNewChart, p.NewChart)
	p.TelegramClient.CreateHandler("/new_chart", p.ValidateNewChart)
	p.TelegramClient.CreateHandler("/chart", p.GenerateChart)

	// Configuration menu
	p.TelegramClient.CreateHandler(&btnConfiguration, func(c telebot.Context) error {
		return c.Send("What do you want to do ?", configurationMenu)
	})
	p.TelegramClient.CreateHandler("/export_db", p.ExportDB)
	p.TelegramClient.CreateHandler(&btnExportDB, p.ExportDB)
	p.TelegramClient.CreateHandler("/reload_config", p.ReloadConfigFile)
	p.TelegramClient.CreateHandler(&btnReloadConfig, p.ReloadConfigFile)
	p.TelegramClient.CreateHandler("/live_config", p.ShowLiveConfig)
	p.TelegramClient.CreateHandler(&btnShowLiveConfig, p.ShowLiveConfig)
	p.TelegramClient.CreateHandler("/config_file", p.ShowConfigFile)
	p.TelegramClient.CreateHandler(&btnShowConfigFile, p.ShowConfigFile)
	p.TelegramClient.CreateHandler("/list_coins", p.ListCoins)
	p.TelegramClient.CreateHandler(&btnListCoins, p.ListCoins)
	p.TelegramClient.CreateHandler(&btnEditCoins, p.EditCoins)
	p.TelegramClient.CreateHandler("/edit_coins", p.ValidateCoinEdit)
	p.TelegramClient.CreateHandler(&btnEditJump, p.EditJump)
	p.TelegramClient.CreateHandler("/edit_jump", p.ValidateJumpEdit)

	// Notifications menu
	p.TelegramClient.CreateHandler(&btnNotification, p.Notification)
	p.NotificationDebug(ctx)
	p.NotificationWarn(ctx)
	p.NotificationInfo(ctx)
	p.NotificationError(ctx)

	// Setup menus
	mainMenu.Reply(
		mainMenu.Row(btnBalance, btnLast10Jumps),
		mainMenu.Row(btnNextJump, btnChart),
		mainMenu.Row(btnConfiguration),
	)
	configurationMenu.Reply(
		configurationMenu.Row(btnEditCoins, btnListCoins),
		configurationMenu.Row(btnEditJump, btnReloadConfig),
		configurationMenu.Row(btnShowLiveConfig, btnShowConfigFile),
		configurationMenu.Row(btnNotification, btnExportDB, btnBackToMainMenu),
	)
	notificationMenu.Reply(
		notificationMenu.Row(btnDebug, btnInfo, btnWarn, btnError),
		notificationMenu.Row(btnBackToMainMenu),
	)

	p.TelegramClient.CreateHandler("/menu", func(c telebot.Context) error {
		return c.Send("What do you want to do ?", mainMenu)
	})
}

func (p *Handlers) Help(c telebot.Context) error {
	parts := []string{
		"Available commands are:",
	}

	parts = append(parts, availableCommands...)

	return c.Send(strings.Join(parts, "\n"), telebot.ModeHTML)
}

func (p *Handlers) BackToMainMenu(c telebot.Context) error {
	return c.Send("Back to menu", mainMenu)
}
