package handlers

import (
	"context"
	"github.com/erwanlbp/trading-bot/pkg/binance"
	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/repository"
	"github.com/erwanlbp/trading-bot/pkg/telegram"
)

type Handlers struct {
	Logger                  *log.Logger
	TelegramClient          *telegram.Client
	BinanceClient           *binance.Client
	NotificationLevelConfig []string
	Repository              *repository.Repository
}

func NewHandlers(l *log.Logger, c *telegram.Client, b *binance.Client, r *repository.Repository, n []string) *Handlers {
	return &Handlers{
		Logger:                  l,
		TelegramClient:          c,
		BinanceClient:           b,
		NotificationLevelConfig: n,
		Repository:              r,
	}
}

func (p *Handlers) InitHandlers(ctx context.Context) {
	p.InitMenu(ctx)
}
