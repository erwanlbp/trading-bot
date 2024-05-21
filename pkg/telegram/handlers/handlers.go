package handlers

import (
	"context"
	"fmt"

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
	BinanceClient  binance.Interface
	Repository     *repository.Repository
	GlobalConf     globalconf.GlobalConfModifier
}

func NewHandlers(l *log.Logger, conf *configfile.ConfigFile, c *telegram.Client, b binance.Interface, r *repository.Repository, gc globalconf.GlobalConfModifier) *Handlers {
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
	p.InitMenu(ctx)
}

func (p *Handlers) CreatePaginatedHandlers(messagePaginated map[interface{}]string, defaultValue interface{}, selector *telebot.ReplyMarkup) []telebot.Btn {
	if len(messagePaginated) < 2 {
		return []telebot.Btn{}
	}

	var buttons []telebot.Btn
	for i := range messagePaginated {
		btn := telebot.Btn{
			Unique: fmt.Sprint(i),
			Text:   "Show " + fmt.Sprint(i) + " value",
			Data:   fmt.Sprint(i),
		}
		buttons = append(buttons, btn)
	}

	for i := range buttons {
		b := buttons[i]
		p.TelegramClient.CreateHandler(&b, func(c telebot.Context) error {
			// TODO not display current inline button but loosing ref to handler ?
			//selector := &telebot.ReplyMarkup{}
			//buttonsWithoutCurrent := util.FilterSlice(buttons, func(btn telebot.Btn) bool {
			//	return btn.Unique != defaultValue
			//})
			//selector.Inline(selector.Row(buttonsWithoutCurrent...))
			return c.Edit(telegram.FormatForMD(messagePaginated[b.Unique]), selector)
		})
	}

	//buttonsWithoutDefault := util.FilterSlice(buttons, func(btn telebot.Btn) bool {
	//	return btn.Unique != defaultValue
	//})
	return buttons
}
