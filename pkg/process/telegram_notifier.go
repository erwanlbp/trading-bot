package process

import (
	"context"
	"github.com/erwanlbp/trading-bot/pkg/util"

	"github.com/erwanlbp/trading-bot/pkg/eventbus"
	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/telegram"
)

type TelegramNotifier struct {
	Logger         *log.Logger
	EventBus       *eventbus.Bus
	TelegramClient *telegram.Client
}

func NewTelegramNotifier(l *log.Logger, e *eventbus.Bus, c *telegram.Client) *TelegramNotifier {
	return &TelegramNotifier{
		Logger:         l,
		EventBus:       e,
		TelegramClient: c,
	}
}

func (n TelegramNotifier) Start(ctx context.Context) {
	sub := n.EventBus.Subscribe(eventbus.SendNotification)

	go sub.Handler(ctx, n.SendNotification)
}

func (n TelegramNotifier) SendNotification(ctx context.Context, e eventbus.Event) {
	n.TelegramClient.Send(util.ToJSON(e.Payload))
}
