package handlers

import (
	"context"
	"github.com/erwanlbp/trading-bot/pkg/binance"
	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/telegram"
	"gopkg.in/telebot.v3"
)

type Handlers struct {
	Logger                  *log.Logger
	TelegramClient          *telegram.Client
	BinanceClient           *binance.Client
	NotificationLevelConfig []string
}

func NewHandlers(l *log.Logger, c *telegram.Client, b *binance.Client, n []string) *Handlers {
	return &Handlers{
		Logger:                  l,
		TelegramClient:          c,
		BinanceClient:           b,
		NotificationLevelConfig: n,
	}
}

func (p *Handlers) Start(ctx context.Context) {
	go func() {
		commands := telebot.CommandParams{
			Commands:     []telebot.Command{MenuHandler, BalanceHandler, NotificationsHandler},
			LanguageCode: "fr",
		}
		p.TelegramClient.SetCommands(commands)
		p.InitHandlers(ctx)
		// TODO: how to close properly ?
		p.TelegramClient.StartBot()
	}()
}

func (p *Handlers) InitHandlers(ctx context.Context) {
	p.Menu(ctx)
	p.Balance(ctx, p.BinanceClient)
	p.NotificationLevel(ctx, p.NotificationLevelConfig)
}
