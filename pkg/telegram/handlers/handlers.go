package handlers

import (
	"context"
	"strconv"

	"go.uber.org/zap"
	"gopkg.in/telebot.v3"

	"github.com/erwanlbp/trading-bot/pkg/binance"
	"github.com/erwanlbp/trading-bot/pkg/config/configfile"
	"github.com/erwanlbp/trading-bot/pkg/config/globalconf"
	"github.com/erwanlbp/trading-bot/pkg/log"
	"github.com/erwanlbp/trading-bot/pkg/repository"
	"github.com/erwanlbp/trading-bot/pkg/telegram"
)

type Handlers struct {
	Logger         *log.Logger
	Conf           *configfile.ConfigFile
	TelegramClient *telegram.Client
	BinanceClient  *binance.Client
	Repository     *repository.Repository
	GlobalConf     globalconf.GlobalConfModifier
}

func NewHandlers(l *log.Logger, conf *configfile.ConfigFile, c *telegram.Client, b *binance.Client, r *repository.Repository, gc globalconf.GlobalConfModifier) *Handlers {
	return &Handlers{
		Logger:         l,
		Conf:           conf,
		TelegramClient: c,
		BinanceClient:  b,
		Repository:     r,
		GlobalConf:     gc,
	}
}

func (p *Handlers) InitHandlers(ctx context.Context) {
	p.InitMenu(ctx, p.Conf)
}

func (p *Handlers) CreatePaginatedHandlers(messagePaginated map[int]string, selector *telebot.ReplyMarkup) []telebot.Btn {
	buttons := make([]telebot.Btn, len(messagePaginated))
	for i := range messagePaginated {
		btn := telebot.Btn{
			Unique: strconv.Itoa(i),
			Text:   strconv.Itoa(i),
			Data:   strconv.Itoa(i),
		}
		buttons[i] = btn
		p.TelegramClient.CreateHandler(&btn, func(c telebot.Context) error {
			index, err := strconv.Atoi(btn.Data)
			if err != nil {
				p.Logger.Error("Error while getting index of page to display diff : ", zap.Error(err))
				return c.Edit(messagePaginated[0], selector)
			}
			return c.Edit(messagePaginated[index], selector)
		})
	}
	return buttons
}
