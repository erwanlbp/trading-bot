package handlers

import (
	"context"
	"gopkg.in/telebot.v3"
)

var (
	mainMenu          = &telebot.ReplyMarkup{ResizeKeyboard: true}
	btnBalance        = mainMenu.Text("⚖️ Balances")
	btnLast10Jumps    = mainMenu.Text("Last 10 jumps")
	btnNextJump       = mainMenu.Text("⤴️ Next jump")
	btnConfiguration  = mainMenu.Text("⚙️ Configuration")
	btnBackToMainMenu = mainMenu.Text("⬅️ Back")
	mainRow           = mainMenu.Row(btnBalance, btnLast10Jumps, btnNextJump)
	mainRow2          = mainMenu.Row(btnConfiguration)
)

func (p *Handlers) InitMenu(ctx context.Context) {
	p.Balance(ctx)
	p.LastTenJumps(ctx)
	p.NextJump(ctx)
	p.Configuration(ctx)
	p.Notification(ctx)
	p.BackToMainMenu(ctx)

	// Setup menus
	mainMenu.Reply(mainRow, mainRow2)
	configurationMenu.Reply(configurationRow)

	p.TelegramClient.CreateHandler("/menu", func(c telebot.Context) error {
		return c.Send("What do you want to do ?", mainMenu)
	})
}

func (p *Handlers) BackToMainMenu(ctx context.Context) {
	p.TelegramClient.CreateHandler(&btnBackToMainMenu, func(c telebot.Context) error {
		return c.Send("Back to main menu", mainMenu)
	})
}
