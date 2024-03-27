package process

import (
	"context"
	"encoding/json"
	"github.com/erwanlbp/trading-bot/pkg/eventbus"
	"github.com/erwanlbp/trading-bot/pkg/eventbus/eventdefinition"
	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/telegram"
)

type Notification struct {
	Logger         *log.Logger
	EventBus       *eventbus.Bus
	TelegramClient *telegram.Client
}

func NewNotification(l *log.Logger, e *eventbus.Bus, c *telegram.Client) *Notification {
	return &Notification{
		Logger:         l,
		EventBus:       e,
		TelegramClient: c,
	}
}

func (n Notification) Start(ctx context.Context) {
	sub := n.EventBus.Subscribe(eventbus.SendNotification)

	go sub.Handler(ctx, n.SendNotification)
}

func (n Notification) SendNotification(ctx context.Context, e eventbus.Event) {
	var event eventdefinition.EventNotification
	marshal, err := json.Marshal(e.Payload)
	if err != nil {
	}
	err = json.Unmarshal(marshal, &event)
	if err != nil {
		return
	}

	message := eventdefinition.MapLevelToIcon(event.Level) + " " + event.Message

	n.TelegramClient.Send(message)
}
